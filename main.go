package main

import (
	"fmt"
	"strings"
	"strconv"
	"bufio"
	"os"
	"time"
	"math/rand"
)

const N, S, E, W = -16, 16, 1, -1
const A8, H8, A1, H1 = 0, 7, 112, 119
const E8, E1 = 4, 116
const CASTLE = E * 8
const PIECE = " .pPnNbBrRqQkK"
const RANK = "87654321"
const FILE = "abcdefgh"
var ADVANCES = [...]int{S,N}

func T(a,b int) int{
	return (a + (b * 0x10000))
}

var material = []int{ T(0,0), T(31,85), T(276,215), T(308,209), T(410,320), T(808,612), T(0,0), }
var passerFile = []int{ T(0,0), T(52,-6), T(41,-15), T(48,-36), T(43,-44), T(55,-54), T(88,-62), T(86,-49), T(67,-51), T(0,0), }
var shield = []int { T(0,0), T(22,2), T(36,6), T(-7,22), T(12,12), T(10,11), T(8,10), T(22,3), T(35,-7), T(0,0),}
var tempo int = T(32,25) 
/*

[T(-55.535,53.804), T(-53.307,50.922), T(-50.354,71.25), T(-47.704,91.68), T(-53.684,157.666), T(-14.853,230.961), T(0.0,-0.0), T(-0.0,-0.0), T(0.0,0.0), T(-0.0,-0.0), ]
[T(-6.72,-15.267), ]
[T(-7.191,-21.868), ]
[T(4.899,4.008), ]
[T(-14.451,-3.918), ]

Piece weights
[T(-0.0,0.0), T(-0.092,-0.007), T(-0.088,0.048), T(-0.075,0.041), T(-0.079,0.036), T(-0.039,0.061), T(0.156,0.124), ]
[T(0.0,0.0), T(-0.42,-0.602), T(-0.593,-0.727), T(-0.367,-0.426), T(0.057,0.431), T(0.053,0.845), T(-0.132,0.108), ]
[T(0.114,-0.01), ]
[T(0.218,0.107), ]

Grid weights
[T(74.03,37.805), ]
[T(0.0,-0.0), T(87.305,194.766), T(-38.408,152.24), T(-204.58,288.475), T(-366.913,204.612), T(-725.963,-1039.029), T(-1140.706,596.005), ]
[T(-67.624,-2.464), ]

Risk weights
[T(-0.759,-0.553), T(-0.801,0.025), T(-0.694,0.346), T(-0.627,0.561), T(-0.313,0.687), T(-0.072,0.788), T(0.115,0.915), T(0.129,1.134), T(0.436,0.763), ]

Base Mob Weights
[T(135.039,106.475), T(149.675,83.195), T(59.336,74.376), T(91.425,76.869), T(53.029,73.799), T(64.904,87.605), T(42.164,83.064), T(-35.633,125.568), ]
[T(6.865,73.25), T(-78.966,76.897), T(-73.013,73.322), T(-20.155,61.778), T(-80.371,67.431), T(-35.495,76.897), T(-49.303,76.79), T(-15.538,91.239), ]
[T(-54.653,52.648), T(-77.425,61.825), T(-64.439,46.088), T(-79.872,67.232), T(-77.673,71.929), T(-118.427,65.567), T(-113.19,79.843), T(-60.409,65.151), ]
[T(-38.749,41.805), T(-71.644,48.394), T(-63.749,64.684), T(-90.221,79.684), T(-90.191,73.102), T(-59.094,62.775), T(-56.523,61.365), T(-38.282,67.286), ]
[T(-23.735,62.724), T(-23.999,58.154), T(-77.387,72.513), T(-103.941,74.882), T(-88.659,65.657), T(-77.747,67.073), T(-54.633,69.753), T(-25.438,64.3), ]
[T(-25.02,76.037), T(-61.686,83.608), T(-83.584,72.537), T(-84.161,68.223), T(-102.671,74.209), T(-81.809,73.169), T(-116.283,81.163), T(-55.528,65.635), ]
[T(-23.331,67.779), T(-81.21,73.968), T(-58.581,58.975), T(-72.301,78.812), T(-50.572,66.57), T(-84.041,57.272), T(-105.611,80.949), T(-20.549,84.338), ]
[T(-54.074,68.184), T(-3.276,72.66), T(-24.928,68.409), T(-18.225,80.236), T(-28.511,73.845), T(8.775,68.703), T(5.117,63.698), T(51.871,-14.391), ]
*/

var phaseWeights = [14]int{0,0,0,0,1,1,1,1,2,2,4,4,0,0}

var Zobrist [16][128]uint64

var nodes int = 0


type Board struct {
	squares    [128]int8
	kings      [2]int
	enpassant  int
	zobrist    uint64
	mobilities [2]int
	phase int
	sidetomove int8
}

type Move struct {
	start int8
	end   int8
}
var nullmove Move = Move{9,9}

func (move Move) stringify() string {
	return string(FILE[move.start&7]) + string(RANK[move.start>>4]) + string(FILE[move.end&7]) + string(RANK[move.end>>4])
}

func Parse(sqstr string) int {
	return strings.Index(FILE, sqstr[0:1]) + ((strings.Index(RANK, sqstr[1:2])) * 16)
}

func FromFen(fen string) Board {
	fenparts := strings.Split(fen, " ")
	board := Board{}
	i := 0
	for n := range fenparts[0] {
		char := fenparts[0][n : n+1]
		piece := int8(strings.Index(PIECE, char))
		if piece >= 0 {
			board.Edit(i, piece)
			if piece == 12 {
				board.kings[0] = i
			} else if piece == 13 {
				board.kings[1] = i 
			}
		} else if char == "/" {
			i += 7
		} else {
			i += 7 - strings.Index(RANK, char)
		}
		i += 1
	}

	if fenparts[1] == "w" {
		board.sidetomove = 1
	}

	if strings.Index(fenparts[2], "k") > -1 {
		board.Edit(H8+CASTLE, 1)
	}
	if strings.Index(fenparts[2], "q") > -1 {
		board.Edit(A8+CASTLE, 1)
	}
	if strings.Index(fenparts[2], "K") > -1 {
		board.Edit(H1+CASTLE, 1)
	}
	if strings.Index(fenparts[2], "Q") > -1 {
		board.Edit(A1+CASTLE, 1)
	}

	if fenparts[3] != "-" {
		board.enpassant = Parse(fenparts[3])
	}

	return board
}

func (board *Board) Edit(index int, newpiece int8) {
	oldpiece := board.squares[index]
	board.phase += phaseWeights[newpiece]
	board.phase -= phaseWeights[oldpiece]
	board.squares[index] = newpiece
	board.zobrist ^= Zobrist[oldpiece][index]
	board.zobrist ^= Zobrist[newpiece][index]
}

var rays = [7]bool{false, false, false, true, true, true, false}

var patterns = [7][]int{
	{},
	{},
	{N + N + W, N + N + E, S + S + W, S + S + E, E + E + N, E + E + S, W + W + N, W + W + S},
	{N + W, N + E, S + W, S + E},
	{N, S, E, W},
	{N, S, E, W, N + W, N + E, S + W, S + E},
	{N, S, E, W, N + W, N + E, S + W, S + E},
}

func (board *Board) IsHomeRow(i int) bool {
	if board.sidetomove == 1 {
		return A1+N <= i
	}
	return i <= H8+S
}

func (board *Board) GenerateLegalMoves(capturesOnly bool) []Move {
	//weights := [128]int{}

	/*
	j := 0
	if (board.sidetomove == 0) {
		j = 120
	}

	for i := range 128 {
		weights[i^j] = 
	}*/


	









	moves := []Move{}

	var ourPawn int8 = 2 + board.sidetomove
	advance := ADVANCES[board.sidetomove]

	mobility := 0

	for i, piece := range board.squares {
		if piece < 2 || piece & 1 != board.sidetomove {
			continue
		}
		piecetype := piece / 2

		if piecetype == 1 {
			if !capturesOnly && board.squares[i+advance] == 0 {
				moves = append(moves, Move{int8(i), int8(i + advance)})
				if board.IsHomeRow(i) && board.squares[i+advance+advance] == 0 {
					moves = append(moves, Move{int8(i), int8(i + advance + advance)})
				}
			}

			for _, pawnCapture := range []int{i + advance + W, i + advance + E} {
				if pawnCapture&0x88 == 0 {
					victim := board.squares[pawnCapture]
					if victim != 0 && (victim&1 != piece&1) {
						moves = append(moves, Move{int8(i), int8(pawnCapture)})
					}
				}
			}
		} else {
			ray := rays[piecetype]
			pattern := patterns[piecetype]
			//mobValue := mobChart[piecetype]


			for _, dir := range pattern {
				for end := i + dir; (end & 0x88) == 0; end += dir {
					victim := board.squares[end]

					if victim != 0 {
						if victim&1 != piece&1 {
							moves = append(moves, Move{int8(i), int8(end)})
							//mobility += mobValue
						}
					} else {
						//mobility += mobValue
						if !capturesOnly {
							moves = append(moves, Move{int8(i), int8(end)})
						}
					}

					if victim != 0 || !ray {
						break
					}
				}
			}
		}
	}

	if board.enpassant != 0 {
		for _, enpassantStart := range []int{board.enpassant - advance + W, board.enpassant - advance + E} {
			if board.squares[enpassantStart] == ourPawn {
				moves = append(moves, Move{int8(enpassantStart), int8(board.enpassant)})
			}
		}
	}



	kingIndex := board.kings[board.sidetomove]
	if !capturesOnly && (kingIndex == E8 || kingIndex == E1) && !board.attacked(kingIndex, 1-board.sidetomove) {
		if board.squares[kingIndex+E+E+E+CASTLE] == 1 && board.squares[kingIndex+E+E] == 0 && board.squares[kingIndex+E] == 0 {
			moves = append(moves, Move{int8(kingIndex), int8(kingIndex+E+E)})
		} 
		if board.squares[kingIndex+W+W+W+W+CASTLE] == 1 && board.squares[kingIndex+W+W+W] == 0 && board.squares[kingIndex+W+W] == 0 && board.squares[kingIndex+W] == 0 {
			moves = append(moves, Move{int8(kingIndex), int8(kingIndex+W+W)})
		}
	}

	board.mobilities[board.sidetomove] = mobility

	return moves
}

func (board *Board) attacked(start int, attacker int8) bool {
	advance := ADVANCES[attacker]

	for _, dir := range []int{W - advance, E - advance} {
		if ((start + dir) & 0x88) != 0 {
			continue
		}
		if (board.squares[start + dir] == 2 + attacker) {
			return true
		}
	}

	for i, dir := range []int{N,S,E,W,N+W,N+E,S+E,S+W} {
		for sq := start + dir; (0x88 & sq) == 0; sq += dir {
			piece := board.squares[sq]
			if piece != 0 {
				if (i < 4 && (piece == 10 + attacker  || piece == 8 + attacker)) {
					return true
				}
				if (i >= 4 && (piece == 10 + attacker || piece == 6 + attacker)) {
					return true
				}
				break
			}
		}
	}

	for i, dir := range []int{N,S,E,W,N+W,N+E,S+E,S+W, N + N + W, N + N + E, S + S + W, S + S + E, E + E + N, E + E + S, W + W + N, W + W + S} {
		if ((start + dir) & 0x88) != 0 {
			continue
		}
		piece := board.squares[start + dir]

		if (i >= 8 && (piece == 4 + attacker)) {
			return true
		}
		if (i < 8 && (piece == 12 + attacker)) {
			return true
		}
	}

	return false
}

func (board *Board) Apply(move Move) *Board {
	// TODO: see perf difference vs just using non pointer method

	currentBoard := *board
	copyBoard := currentBoard

	movingPiece := copyBoard.squares[move.start]
	copyBoard.Edit(int(move.start), 0)
	copyBoard.Edit(int(move.end), movingPiece)

	advance := ADVANCES[copyBoard.sidetomove]
	newEP := 0

	if movingPiece <= 3 {
		// Enpassant capture -> remove extra pawn
		if int(move.end) == copyBoard.enpassant {
			copyBoard.Edit(int(move.end)-advance, 0)
		}
		
		// Double push -> set enpassant square
		if int(move.end-move.start) == advance+advance {
			newEP = int(move.start) + advance
		}

		// Promotion -> turn it into a queen
		if move.end < A8+S || move.end > H1+N {
			copyBoard.Edit(int(move.end), 10 + copyBoard.sidetomove)
		}
	} else if movingPiece >= 12 {
		// Update kingpos
		copyBoard.kings[copyBoard.sidetomove] = int(move.end)
		
		// Castle -> move rook, do some validation
		if move.end-move.start == W+W {
			if copyBoard.attacked(int(move.start) + W, 1 - copyBoard.sidetomove) {
				return nil
			}
			copyBoard.Edit(int(move.end) + W + W, 0)
			copyBoard.Edit(int(move.end) + E, 8+copyBoard.sidetomove)
		}
		if move.end-move.start == E+E {
			if copyBoard.attacked(int(move.start) + E, 1 - copyBoard.sidetomove) {
				return nil
			}
			copyBoard.Edit(int(move.end) + E, 0)
			copyBoard.Edit(int(move.end) + W, 8+copyBoard.sidetomove)
		}

		// Invalidate castling
		//fmt.Println(move.end)

		if copyBoard.sidetomove == 1 {
			copyBoard.Edit(A1 + CASTLE, 0)
			copyBoard.Edit(H1 + CASTLE, 0)

		} else {
			copyBoard.Edit(A8 + CASTLE, 0)
			copyBoard.Edit(H8 + CASTLE, 0)
		}
	}

	kingIndex := copyBoard.kings[copyBoard.sidetomove]

	if copyBoard.attacked(kingIndex, 1 - copyBoard.sidetomove) {
		return nil
	}

	copyBoard.sidetomove = 1 - copyBoard.sidetomove
	copyBoard.enpassant = newEP

	if copyBoard.squares[int(move.start)+CASTLE] != 0 {
		copyBoard.Edit(int(move.start)+CASTLE, 0)
	}
	if copyBoard.squares[int(move.end)+CASTLE] != 0 {
		copyBoard.Edit(int(move.end)+CASTLE, 0)
	}

	return &copyBoard
}

func (board *Board) print(){
	for i, piece := range board.squares {
		fmt.Print(string(PIECE[piece]))
		if i % 16 == 15 {
			fmt.Println("")
		}
	}
	fmt.Println(board.zobrist)
}

func (board *Board) Hash() uint64 {
	return board.zobrist ^ Zobrist[15][board.enpassant] ^ Zobrist[15][120+board.sidetomove]
}

func perft(perftboard *Board, depth int, maxdepth int) int {
    if (depth == 0) { return 1 }

    movelist := perftboard.GenerateLegalMoves(false)
    nodes := 0

    for i, move := range movelist {
        nextBoard := perftboard.Apply(move)
        if (nextBoard != nil) { 
        	subnodes := perft(nextBoard, depth - 1, maxdepth) 
        	nodes += subnodes 
        	if depth == maxdepth {
        		fmt.Println(i, move.stringify(), subnodes)
        	}
        }
    }

    return nodes
}

func findAfter(word string, strlist []string) []string {
	for i, v := range strlist {
		if v == word {
			return strlist[i+1:]
		}
	}
	return []string{}
}

func eval(board *Board) int {
	score := board.mobilities[1] - board.mobilities[0]

	whiterear := [10]int{0,0,0,0,0,0,0,0,0,0,}
	blackrear := [10]int{7,7,7,7,7,7,7,7,7,7,}

	for sq, piece := range board.squares {
		
		if piece & 1 == 1 {
			score += material[piece / 2]
		} else {
			score -= material[piece / 2]
		}


		if piece / 2 == 1 {
			pawnfile := (sq & 7) + 1
			pawnrank := sq >> 4

			if (piece & 1) == 1 {
				whiterear[pawnfile] = max(whiterear[pawnfile], pawnrank)
			} else {
				blackrear[pawnfile] = min(blackrear[pawnfile], pawnrank)
			}
		}
	}

	/*
	wkingfile := (board.kings[1]&7) + 1
	wkingrank := board.kings[1] >> 4
	bkingfile := (board.kings[0]&7) + 1
	bkingrank := board.kings[0] >> 4

	for i := -1; i <= 1; i++ {
		if whiterear[wkingfile + i] != 0 && whiterear[wkingfile + i] < wkingrank {
			score += shieldChart[wkingfile + i]
		}
		if blackrear[bkingfile + i] != 7 && blackrear[bkingfile + i] > bkingrank {
			score -= shieldChart[bkingfile + i]
		}
	}

	for sq, piece := range board.squares {
		if piece / 2 == 1 {
			pfile := (sq & 7) + 1
			prank := sq >> 4

			if piece & 1 == 1 {
				if whiterear[pfile-1] == 0 && whiterear[pfile+1] == 0 {
					score -= 20
				}
				if blackrear[pfile - 1] >= prank && blackrear[pfile] >= prank && blackrear[pfile + 1] >= prank {
					score += 30

				}

			} else {
				if blackrear[pfile-1] == 7 && blackrear[pfile+1] == 7 {
					score += 20
				}
				if whiterear[pfile - 1] <= prank && whiterear[pfile] <= prank && whiterear[pfile + 1] <= prank {
					score -= 30
				}
			}
		}
	}

*/

	eg := (score + 0x8000) >> 16;
	mg := int(int16(score))

	score = ((mg * board.phase) + (eg * (24-board.phase)))/24

	if board.sidetomove == 0 {
		return -score + 20
	}

	return score + 20
}


type entry struct {
	key uint64
	move Move
	depth int
	score int
	bound int8
}

const hashsize = 16777216
var table [hashsize]entry

var history [14][128]int

func alphabeta(board *Board, alpha, beta, depth, ply int, nullallowed bool) int {
	nodes += 1
	bestScore := -9999 + ply

	// standpat
	staticEval := eval(board)
	if (depth <= 0) {
		bestScore = staticEval
		if bestScore > alpha {
			alpha = bestScore
		}
		if bestScore >= beta {
			return bestScore
		}
	}

	moves := board.GenerateLegalMoves(depth <= 0)
	priorities := make([]int, len(moves), len(moves))

	hash := board.Hash()

	if ply != 0 {
		for _, rep := range repetition {
			if rep == hash {
				return 0
			}
		}
	}


	tt := table[hash % hashsize]

	if ply != 0 {
		if tt.key == hash {
			if tt.depth >= depth || 0 >= depth {
				if (tt.bound == 1 && tt.score <= alpha) {return tt.score}
				if (tt.bound == -1 && tt.score >= beta) {return tt.score}
				if (tt.bound == 0) {return tt.score}
			}
		} else {
			depth -= 1
		}
	}

	pv := beta - alpha != 1
	if ply != 0 && depth > 0 && !pv {
		// Null move pruning NMP
		if staticEval >= beta && nullallowed && depth > 3 {
			nmBoard := board.Apply(nullmove)
			if nmBoard != nil {
				nmScore := alphabeta(nmBoard, -beta, -beta+1, 3+depth/6, ply+1, false)
				if nmScore >= beta {
					return beta
				}
			}
		}

		// Reverse futility pruning
		if (depth < 4 && staticEval - depth * 75 > beta) { return staticEval }
	}




	// Prunings should only happen above this
	repetition = append(repetition, hash)

	for i, move := range moves {
		if board.squares[move.end] != 0 {
			priorities[i] = (int(board.squares[move.end]) * 20) - int(board.squares[move.start]) + 10000
		} else {
			priorities[i] = history[board.squares[move.start]][move.end]
		}
		if tt.move.end == move.end && tt.move.start == move.start {
			priorities[i] = 100000
		}
	}

	legals := 0
	var bestMove Move
	var boundtype int8 = -1

	for i := range moves {
		// Selection sort
		besti := i
		for j:=i;j<len(moves);j++ {
			if priorities[besti] < priorities[j] {
				besti = j
			}
		}

		nextMove := moves[besti]

		priorities[i], priorities[besti] = priorities[besti], priorities[i]
		moves[i], moves[besti] = moves[besti], moves[i]



		nextBoard := board.Apply(nextMove)
		if nextBoard == nil {
			continue
		}

		legals += 1

		reduction := 0
		if legals > 4 {
			reduction = legals/16 + depth / 8
		}

		if !pv {
			reduction += 1
		}

		
		var score int

		if ((legals == 1 || depth <= 0)){
			score = -alphabeta(nextBoard, -beta, -alpha, depth - 1, ply + 1, true)
		} else {
			score = -alphabeta(nextBoard, -alpha-1, -alpha, depth - 1 - reduction, ply + 1, true)
			if score > alpha {
				score = -alphabeta(nextBoard, -alpha-1, -alpha, depth - 1, ply + 1, true)
				if score > alpha {
					score = -alphabeta(nextBoard, -beta, -alpha, depth - 1, ply+1, true)
				}
			}
		}

		if score > bestScore {
			bestScore = score
			bestMove = nextMove
		}

		if score > alpha {
			boundtype = 0
			alpha = score
		}

		if score >= beta {
			boundtype = 1

			if (board.squares[nextMove.end] == 0) {
				hh := &history[board.squares[nextMove.start]][nextMove.end]

				*hh += depth * depth

				for m:=0;m<i;m++ {
					if moves[i].end == 0 {
						hhm := &history[board.squares[moves[i].start]][moves[i].end]
						*hhm -= depth * depth
					}
				}
			}

			break
		}

	}

	repetition = repetition[:len(repetition)-1]

	if legals == 0 && depth > 0 {
		// TODO: stalemate
		return -9999
	}

	if bestMove.start != bestMove.end {
		table[hash % hashsize] = entry{hash, bestMove, depth, bestScore, boundtype}
	}

	return bestScore
}

var uciBoard Board
var repetition []uint64 = []uint64{}

func printpv() string {
	pvstr := ""
	board := &uciBoard
	for range 20 {
		if board == nil {
			break
		}
		m := table[board.Hash() % hashsize].move
		if board.Hash() != table[board.Hash() % hashsize].key {
			break
		}

		pvstr += m.stringify() + " "

		//fmt.Println(m.stringify(), table[board.Hash() % hashsize].depth)
		board = board.Apply(m)
	}
	return strings.TrimSpace(pvstr)
}

func parseuci(line string) {
	args := strings.Fields(line)

	switch string(args[0]) {
	case "uci":
		fmt.Println("id name Tabby")
		fmt.Println("id author ffloof")
		fmt.Println("uciok")
	case "isready":
		fmt.Println("readyok")
	case "print":

		uciBoard.print()
	case "perft":
		start := time.Now().UnixMilli()
		fmt.Println("total", perft(&uciBoard, 5, 5))
		fmt.Println("time", time.Now().UnixMilli() - start)

	case "position":
		uciBoard = FromFen(strings.Join(findAfter("fen", args)[0:4], " "))
		for _, movestr := range findAfter("moves", args) {
			repetition = append(repetition, (uciBoard.Hash()))
			uciBoard = *(uciBoard.Apply(Move{int8(Parse(movestr[0:2])), int8(Parse(movestr[2:4]))}))
			// TODO: add underpromotion condition?
		}
	case "go":
		timeAlloc := 1000
		if len(findAfter("movetime", args)) != 0 {
			timeAlloc, _ = strconv.Atoi(findAfter("movetime", args)[0])
		}

		if len(findAfter("wtime", args)) != 0 && uciBoard.sidetomove == 1 {
			timeAlloc, _ = strconv.Atoi(findAfter("wtime", args)[0])
			timeAlloc /= 30
		}

		if len(findAfter("btime", args)) != 0 && uciBoard.sidetomove == 0 {
			timeAlloc, _ = strconv.Atoi(findAfter("btime", args)[0])
			timeAlloc /= 30
		}

		nodes = 0
		start := time.Now().UnixMilli()
		for depth := 1; depth <= 100; depth++ {
			fmt.Println("info score cp", alphabeta(&uciBoard, -10000, 10000, depth, 0, true), "depth", depth, "time", time.Now().UnixMilli() - start, "nodes", nodes, "pv", printpv())

			for i := range 14 {
				for j := range 128 {
					history[i][j] /= 8
				}
			}

			if int(time.Now().UnixMilli() - start) > timeAlloc {
				break
			}

		}
		fmt.Println("bestmove", table[uciBoard.Hash() % hashsize].move.stringify())
	case "quit":
		return
	}
}

func main() {
	fmt.Println("info string Started")
	for i := range 15 {
		for j := range 128 {
			Zobrist[i+1][j] = rand.Uint64()
		}
	}

	uciBoard = FromFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	reader := bufio.NewReader(os.Stdin)

	//parseuci("position fen rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8")

	for {
		line, _ := reader.ReadString('\n')
		line = strings.Replace(line, "startpos", "fen rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 1)
		parseuci(line)
	}
}

// Eval centered around
// 1. Material
//     - static count
//          - perhaps after a threshold is hit we can change to meme eval
// 2. Activity
//     - mobility control map
//          - mobility
//              - penalize squares with enemy pawn guard
//              - scale infront of backwards pawns?
//              - king ring
//          - attacks
//              - bonus on backwards and isolated pawns?
//          - we can also dual purpose this map for move ordering, no need to do seperate mvvlva?
//          - should we flip this based on king sides, perhaps create several distinct maps?
// 3. King Safety
//     - tbh king safety is not that big a deal for engines, but it leads to more fun attacking games
//     - king pawn shield quality
//     - queen tropism?
// 4. Pawn Structure
//     - pawn shield
//     - backwards pawns
//     - isolated pawns
//     - passed pawns
//         - opposite king passer bonus?
//     - encourage trades in completely winning positions? (perhaps implement 50 move and slowly taper eval to 0)

//     - risk scaling, perhaps some combination of activity and/or king safety?

// Still need to figure out how to reward the queen moreso than other pieces for attacking king ring



// Ben finegolds middle name is philip
// Should make a stream where people vote on best move
