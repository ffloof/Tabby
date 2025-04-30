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

func decode(eval, phase int) int {
	eg := (eval + 0x8000) >> 16;
	mg := int(int16(eval))
	return ((mg * phase) + (eg * (24-phase)))/24
}

/*
===
19 0.23942338483352407

Linear terms
{T(0,0), T(42,87), T(297,294), T(345,326), T(407,600), T(867,1129), T(0,0), }
{T(0,0), T(-4,58), T(-19,44), T(-11,26), T(-6,8), T(0,0), T(32,-10), T(12,9), T(1,6), T(0,0), }
{T(0,0), T(28,12), T(43,14), T(0,25), T(13,15), T(15,7), T(17,19), T(15,4), T(37,-8), T(0,0), }
{T(30,39), }
{T(2,-17), T(1,-11), T(-6,11), T(9,52), T(-1,139), T(-13,230), T(0,0), T(0,0), T(0,0), T(0,0), }
{T(0,0), T(34,-5), T(6,-11), T(10,-3), T(9,-1), T(4,5), T(-31,-41), }
{T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(2,-12), T(39,15), T(53,26), T(67,-21), T(60,4), T(162,9), T(0,0), T(-5,24), T(-38,-6), T(28,49), T(42,58), T(35,12), T(117,5), T(0,0), T(2,13), T(15,41), T(-22,11), T(22,27), T(36,23), T(62,91), T(0,0), T(-13,19), T(7,26), T(21,32), T(-83,37), T(63,-24), T(205,-5), T(0,0), T(-1,3), T(-7,25), T(-12,53), T(-12,1), T(12,52), T(46,120), T(0,0), T(73,46), T(-3,42), T(-25,47), T(-173,75), T(-384,-147), T(0,0), }
{T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(7,-3), T(44,18), T(74,54), T(43,38), T(47,-6), T(154,48), T(0,0), T(-7,-3), T(-37,-12), T(29,33), T(91,-3), T(22,35), T(129,19), T(0,0), T(-4,-10), T(16,22), T(-25,18), T(53,39), T(55,34), T(61,64), T(0,0), T(-12,-14), T(-6,-5), T(-5,-16), T(-46,-27), T(70,-20), T(135,13), T(0,0), T(-1,-2), T(-5,-11), T(-19,10), T(15,-41), T(44,53), T(96,36), T(0,0), T(20,5), T(-69,-3), T(-43,7), T(-30,2), T(-67,-119), T(0,0), }
{T(0,0), T(-4,-4), T(-8,-5), T(-3,4), T(-7,1), T(-3,0), T(19,13), }
{T(-15,-7), }
{T(-5,2), }

Mobility weights
{T(62,2), }
{T(20,19), T(12,9), T(16,6), T(5,13), T(-30,19), }

Risk weights
[0.296 0.53  0.723 0.943 0.873 1.024 1.026 1.06  1.   ]

Board Weights
[[ 0.058 -0.176  0.069  0.207  0.402  0.41   0.619  0.882]
 [ 0.42   0.497  0.576  0.291  0.544  0.252  0.655  0.356]
 [ 0.3    0.486  0.593  0.474  0.503  0.654  0.435  0.239]
 [ 0.304  0.474  0.334  0.672  0.599  0.413  0.518  0.427]
 [ 0.102  0.227  0.523  0.668  0.641  0.458  0.427  0.115]
 [ 0.037  0.332  0.503  0.474  0.615  0.406  0.824  0.121]
 [ 0.033  0.28   0.457  0.417  0.344  0.564  0.586  0.111]
 [ 0.364  0.336  0.23   0.336  0.368  0.094  0.223  0.132]]
===
*/

var e_material = []int{T(0,0), T(42,87), T(297,294), T(345,326), T(407,600), T(867,1129), T(0,0), }
var e_passerFile = []int{T(0,0), T(-4,58), T(-19,44), T(-11,26), T(-6,8), T(0,0), T(32,-10), T(12,9), T(1,6), T(0,0), }
var e_shield = []int{T(0,0), T(28,12), T(43,14), T(0,25), T(13,15), T(15,7), T(17,19), T(15,4), T(37,-8), T(0,0), }
var e_tempo int = T(30,39)
var e_passerRank = []int{T(2,-17), T(1,-11), T(-6,11), T(9,52), T(-1,139), T(-13,230), T(0,0), T(0,0), T(0,0), T(0,0), }
var e_kingAttacks = []int{T(0,0), T(34,-5), T(6,-11), T(10,-3), T(9,-1), T(4,5), T(-31,-41), }
var e_attacks = []int{
	T(0,0), T(0,0),    T(0,0),    T(0,0),    T(0,0),     T(0,0),       T(0,0), 
	T(0,0), T(2,-12),  T(39,15),  T(53,26),  T(67,-21),  T(60,4),      T(162,9), 
	T(0,0), T(-5,24),  T(-38,-6), T(28,49),  T(42,58),   T(35,12),     T(117,5), 
	T(0,0), T(2,13),   T(15,41),  T(-22,11), T(22,27),   T(36,23),     T(62,91), 
	T(0,0), T(-13,19), T(7,26),   T(21,32),  T(-83,37),  T(63,-24),    T(205,-5), 
	T(0,0), T(-1,3),   T(-7,25),  T(-12,53), T(-12,1),   T(12,52),     T(46,120), 
	T(0,0), T(73,46),  T(-3,42),  T(-25,47), T(-173,75), T(-384,-147), T(0,0), }

var e_restricted = []int{T(0,0), T(-4,-4), T(-8,-5), T(-3,4), T(-7,1), T(-3,0), T(19,13), }
var e_isolated int = T(-15,-7)
var e_backwards int = T(-5,2)

// Note mobility only counts current square for pawns
var e_mobility = []int{T(0,0), T(62,2), T(20,19), T(12,9), T(16,6), T(5,13), T(-30,19), }


func flip(arr1, arr2 *[128]int, xor int){
	for i := range(len(arr1)) {
		arr2[i] = arr1[i^xor]
	}
}

var e_risk = []int{ 296, 530, 723, 943, 873, 1024, 1026, 1060, 1000 }
var e_table = [2][128]int {
	{},
	{
  58, -176,   69,  207,  402,  410,  619,  882, 0,0,0,0,0,0,0,0,
 420,  497,  576,  291,  544,  252,  655,  356, 0,0,0,0,0,0,0,0,
 300,  486,  593,  474,  503,  654,  435,  239, 0,0,0,0,0,0,0,0,
 304,  474,  334,  672,  599,  413,  518,  427, 0,0,0,0,0,0,0,0,
 102,  227,  523,  668,  641,  458,  427,  115, 0,0,0,0,0,0,0,0,
 037,  332,  503,  474,  615,  406,  824,  121, 0,0,0,0,0,0,0,0,
 033,  280,  457,  417,  344,  564,  586,  111, 0,0,0,0,0,0,0,0,
 364,  336,  230,  336,  368,   94,  223,  132, 0,0,0,0,0,0,0,0,
	}, 
}

const e_divider = 1000

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
	attention := &e_table[board.sidetomove]

	moves := []Move{}

	var ourPawn int8 = 2 + board.sidetomove
	advance := ADVANCES[board.sidetomove]

	mobility := 0

	for i, piece := range board.squares {
		if piece < 2 || piece & 1 != board.sidetomove {
			continue
		}
		piecetype := piece / 2

		mobValue := attention[i]
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
						mobility += decode(e_attacks[piecetype * 7 + (victim/2)], board.phase)
					}
				}
			}
		} else {
			ray := rays[piecetype]
			pattern := patterns[piecetype]
			
			mobValue := 0


			for _, dir := range pattern {
				for end := i + dir; (end & 0x88) == 0; end += dir {
					victim := board.squares[end]

					if victim != 0 {
						if victim&1 != piece&1 {
							mobValue += attention[end]
							mobility += decode(e_attacks[piecetype * 7 + (victim/2)], board.phase)
							moves = append(moves, Move{int8(i), int8(end)})
						}
					} else {
						if !capturesOnly {
							mobValue += attention[end]
							moves = append(moves, Move{int8(i), int8(end)})
						}
					}

					if victim != 0 || !ray {
						break
					}
				}
			}
		}
		mobility += (decode(e_mobility[piecetype],board.phase) * mobValue) / e_divider
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
	score := 0

	whiterear := [10]int{0,0,0,0,0,0,0,0,0,0,}
	blackrear := [10]int{7,7,7,7,7,7,7,7,7,7,}

	for sq, piece := range board.squares {
		
		if piece & 1 == 1 {
			score += e_material[piece / 2]
		} else {
			score -= e_material[piece / 2]
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

	wkingfile := (board.kings[1]&7) + 1
	//wkingrank := board.kings[1] >> 4
	bkingfile := (board.kings[0]&7) + 1
	//bkingrank := board.kings[0] >> 4

	for i := -1; i <= 1; i++ {
		if whiterear[wkingfile + i] != 0 {
			score += e_shield[wkingfile + i]
		}
		if blackrear[bkingfile + i] != 7 {
			score -= e_shield[bkingfile + i]
		}
	}
	/*

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

	score = decode(score, board.phase) 
	score += board.mobilities[1] - board.mobilities[0]
	
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

	flip(&e_table[1], &e_table[0], 112)

	//fmt.Println(e_table)

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
//     - weighted mobility
//          - penalize squares with enemy pawn guard
//          - attacks
//          - should we flip this based on king sides, perhaps create several distinct maps?
// 3. King Safety (covered in other areas)
//     - tbh king safety is not that big a deal for engines, but it leads to more fun attacking games
// 4. Pawn Structure
//     - pawn shield
//     - backwards pawns
//     - isolated pawns
//     - passed pawns
//         - opposite king passer bonus?

// Still need to figure out how to reward the queen moreso than other pieces for attacking king ring


// Ben finegolds middle name is philip
// Should make a stream where people vote on best move
