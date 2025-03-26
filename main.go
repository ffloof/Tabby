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


//[  0  77 257 292 427 933   0]
var matvalues = [...]int{ 0,0,-77,77,-257,257,-292,292,-427,427,-933,933,0,0}

var pieceHit = [7]int{  0,  0,  17,  14,  12,  -3, -11}
var pawnHit = [7]int{ 0,  0, -3,   0,   -1,   0,  29 }
var mobChart = [...]int{0,0,5,2,3,0,0}
var shieldChart = [10]int{ 0, 20, 24, 3, 4, 0, 12, 35, 24, 0}
var pawnChart = [2][128]int{
	{
  0,  0,  0,  0,  0,  0,  0,  0,  0,0,0,0,0,0,0,0,
 10,  0,  0, 10, 10,-10,-10,  0,  0,0,0,0,0,0,0,0,
 10,  0,  0,  0,  0, 10,  0,  0,  0,0,0,0,0,0,0,0,
 10,  0,-10,-20,-20,  0,  0,  0,  0,0,0,0,0,0,0,0,
  0,-10,-10,-30,-30,-20,  0,  0,  0,0,0,0,0,0,0,0,
-50,-50,-40,-40,-40,-50,-30,-30,  0,0,0,0,0,0,0,0,
-90,-90,-90,-70,-70,-80,-80,-80,  0,0,0,0,0,0,0,0,
  0,  0,  0,  0,  0,  0,  0,  0,  0,0,0,0,0,0,0,0,


	},
	{
  0,  0,  0,  0,  0,  0,  0,  0,  0,0,0,0,0,0,0,0,
 90, 90, 90, 70, 70, 80, 80, 80,  0,0,0,0,0,0,0,0,
 50, 50, 40, 40, 40, 50, 30, 30,  0,0,0,0,0,0,0,0,
  0, 10, 10, 30, 30, 20,  0,  0,  0,0,0,0,0,0,0,0,
-10,  0, 10, 20, 20,  0,  0,  0,  0,0,0,0,0,0,0,0,
-10,  0,  0,  0,  0,-10,  0,  0,  0,0,0,0,0,0,0,0,
-10,  0,  0,-10,-10, 10, 10,  0,  0,0,0,0,0,0,0,0,
  0,  0,  0,  0,  0,  0,  0,  0,  0,0,0,0,0,0,0,0,
	},
}

var Zobrist [16][128]uint64

var nodes int = 0


type Board struct {
	squares    [128]int8
	kings      [2]int
	enpassant  int
	sidetomove int8
	zobrist    uint64
	mobilities [2]int
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
			mobValue := mobChart[piecetype]


			for _, dir := range pattern {
				for end := i + dir; (end & 0x88) == 0; end += dir {
					victim := board.squares[end]

					if victim != 0 {
						if victim&1 != piece&1 {
							moves = append(moves, Move{int8(i), int8(end)})
							mobility += mobValue
							if victim / 2 == 1 {
								mobility += pawnHit[piecetype]
							} else {
								mobility += pieceHit[piecetype]
							}
						}
					} else {
						mobility += mobValue
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
		score += matvalues[piece]
		if piece / 2 == 1 {
			score += pawnChart[piece & 1][sq]

			pawnfile := (sq & 7) + 1
			pawnrank := sq >> 4

			if (piece & 1) == 1 {
				whiterear[pawnfile] = max(whiterear[pawnfile], pawnrank)
			} else {
				blackrear[pawnfile] = min(blackrear[pawnfile], pawnrank)
			}
		}
	}

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

	/*
	if piece & 1 == 0:
                if rearpawns[1][pfile - 1] <= prank and rearpawns[1][pfile] <= prank and rearpawns[1][pfile + 1] <= prank:
                    passers[0][pfile] += 1
                if rearpawns[0][pfile-1] > prank and rearpawns[0][pfile+1] > prank:
                    backwards[0][pfile] += 1
                if rearpawns[0][pfile-1] == 12 and rearpawns[0][pfile+1] == 12:
                    isolated[0][pfile] += 1
            else:
                if rearpawns[0][pfile - 1] >= prank and rearpawns[0][pfile] >= prank and rearpawns[0][pfile + 1] >= prank:
                    passers[1][pfile] += 1
                if rearpawns[1][pfile-1] < prank and rearpawns[1][pfile+1] < prank:
                    backwards[1][pfile] += 1
                if rearpawns[1][pfile-1] == 0 and rearpawns[1][pfile+1] == 0:
                    isolated[1][pfile] += 1
*/


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
//     - perhaps after a threshold is hit we can change to meme eval
// 2. King Safety
//     - pawn shield (figure out based on endgames if we should penalize it for king being in front of it)
//     - should implement a scaling "risk" system, that rewards checkmates/kingside attack terms when losing
// 3. Activity
//     - mobility/attacks
//	   - could classify mobility based on if its behind our/opponent pawns
//          - this could be really good for king safety if we just add a kingside condition
// 4. Pawn Structure / Endgames
//     - backwards pawns?
//     - isolated pawns
//     - passed pawns
//     - opposite king passer bonus?
//
//     - encourage trades in completely winning positions (perhaps implement 50 move and slowly taper eval to 0)

// Ben finegolds middle name is philip
// Should make a stream where people vote on best move
