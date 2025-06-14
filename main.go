package main

import (
	"fmt"
	"strings"
	"strconv"
	"bufio"
	"os"
	"time"
	"math/rand"
	"math"
)

import ("flag"
	"runtime/pprof"
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
	return (a + (b * 0x100000000))
}

func decode(eval, phase int) int {
	eg := (eval + 0x80000000) >> 32;
	mg := int(int16(eval))
	return ((mg * phase) + (eg * (24-phase)))/24
}

/*
===
19 0.3027535860054364

Linear terms
{T(0,0), T(25,73), T(161,292), T(185,298), T(225,541), T(505,1002), T(0,0), }
{T(0,0), T(-9,-18), T(-17,-13), T(-16,3), T(-12,28), T(4,93), T(10,145), T(0,0), }
{T(0,0), T(21,3), T(32,4), T(9,10), T(3,8), T(11,8), T(12,11), T(29,6), T(32,-7), T(0,0), }
{T(21,20), }
{T(0,0), T(25,3), T(7,10), T(-3,26), T(-2,42), T(-10,61), T(-3,49), T(-5,46), }
{T(0,0), T(0,0), T(-8,-4), T(-4,-1), T(-7,-5), T(-5,1), T(-9,1), }

{

T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), 
T(0,0), T(0,0), T(44,9), T(41,30), T(56,11), T(58,-21), T(0,0), 
T(0,0), T(-8,5), T(0,0), T(16,19), T(44,15), T(32,-12), T(0,0), 
T(0,0), T(2,7), T(7,25), T(0,0), T(28,22), T(28,102), T(0,0), 
T(0,0), T(-10,10), T(1,16), T(7,11), T(0,0), T(51,19), T(0,0), 
T(0,0), T(1,5), T(2,3), T(-1,19), T(4,4), T(0,0), T(0,0), 
T(0,0), T(37,20), T(-12,27), T(12,25), T(-41,29), T(0,0), T(0,0), }

{T(8,45), }

Mobility weights
{T(0,0), T(-36,-2), T(-20,-13), T(-12,-10), T(-14,-4), T(-5,-12), T(30,-18), 
T(-24,-39), T(-12,0), T(-37,-41), T(-16,-22), }

Risk weights
[0.334 0.629 0.723 1.002 1.118 1.236 1.276 1.145 1.   ]

Board Weights
[[ 0.015  0.016 -0.204 -0.269 -0.269 -0.269 -0.813 -1.169]
 [-0.453 -0.542 -0.519 -0.443 -0.475 -0.466 -0.996 -0.81 ]
 [-0.5   -0.544 -0.696 -0.512 -0.576 -0.818 -0.641 -0.546]
 [-0.379 -0.521 -0.441 -0.556 -0.657 -0.551 -0.525 -0.57 ]
 [-0.212 -0.269 -0.41  -0.487 -0.586 -0.35  -0.308 -0.339]
 [-0.032 -0.351 -0.253 -0.425 -0.437 -0.327 -0.466 -0.273]
 [-0.011 -0.37  -0.303 -0.219 -0.251 -0.492 -0.468 -0.134]
 [-0.168 -0.164 -0.146 -0.185 -0.202 -0.213 -0.218  0.052]]
===
*/

var e_material = []int{T(0,0), T(25,73), T(161,292), T(185,298), T(225,541), T(505,1002), T(0,0), }
var e_passerRank = []int{T(0,0), T(-9,-18), T(-17,-13), T(-16,3), T(-12,28), T(4,93), T(10,145), T(0,0), }
var e_shield = []int{T(0,0), T(21,3), T(32,4), T(9,10), T(3,8), T(11,8), T(12,11), T(29,6), T(32,-7), T(0,0), }
var e_tempo int = T(10,10) //T(21,20)
var e_passerFile = []int{T(0,0), T(25,3), T(7,10), T(-3,26), T(-2,42), T(-10,61), T(-3,49), T(-5,46), }
var e_restricted = []int{T(0,0), T(0,0), T(-8,-4), T(-4,-1), T(-7,-5), T(-5,1), T(-9,1), }

var e_attacks = []int{
	T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), 
	T(0,0), T(0,0), T(44,9), T(41,30), T(56,11), T(58,-21), T(0,0), 
	T(0,0), T(-8,5), T(0,0), T(16,19), T(44,15), T(32,-12), T(0,0), 
	T(0,0), T(2,7), T(7,25), T(0,0), T(28,22), T(28,102), T(0,0), 
	T(0,0), T(-10,10), T(1,16), T(7,11), T(0,0), T(51,19), T(0,0), 
	T(0,0), T(1,5), T(2,3), T(-1,19), T(4,4), T(0,0), T(0,0), 
	T(0,0), T(37,20), T(-12,27), T(12,25), T(-41,29), T(0,0), T(0,0),
}

var e_bishopPair int = T(8,45)

// Note mobility only counts current square for pawns
var e_mobility = []int{T(0,0), T(36,2), T(20,13), T(12,10), T(14,4), T(5,12), T(-30,18), }
var e_phalanxOpen int = T(24,39)
var e_phalanxClosed int = T(12,0)
var e_chainOpen int = T(37,41)
var e_chainClosed int = T(16,22)

func flip(arr1, arr2 *[128]int, xor int){
	for i := range(len(arr1)) {
		arr2[i] = arr1[i^xor]
	}
}

var e_risk = []int{ 334, 629, 723, 1002, 1118, 1236, 1276, 1145, 1000 }
var e_table = [2][128]int {
	{},
{-15, -16, 204, 269, 269, 269, 813, 1169, 0,0,0,0, 0,0,0,0,
 453, 542, 519, 443, 475, 466, 996, 810,  0,0,0,0, 0,0,0,0,
 500, 544, 696, 512, 576, 818, 641, 546,  0,0,0,0, 0,0,0,0,
 379, 521, 441, 556, 657, 551, 525, 570,  0,0,0,0, 0,0,0,0,
 212, 269, 410, 487, 586, 350, 308, 339,  0,0,0,0, 0,0,0,0,
  32, 351, 253, 425, 437, 327, 466, 273,  0,0,0,0, 0,0,0,0,
  11, 370, 303, 219, 251, 492, 468, 134,  0,0,0,0, 0,0,0,0,
 168, 164, 146, 185, 202, 213, 218, -52,  0,0,0,0, 0,0,0,0,},
}

const e_divider = 1000

var phaseWeights = [14]int{0,0,0,0,1,1,1,1,2,2,4,4,0,0}

var Zobrist [16][128]uint64

var nodes int = 0
const MAX_HISTORY = 256

type Board struct {
	squares    [128]int8
	pieceCount [16]int8
	kings      [2]int
	enpassant  int
	zobrist    uint64
	mobilities [2]int
	phase int
	sidetomove int8
	inCheck bool
	ply int
}

type Move struct {
	start int8
	end   int8
}

func (move Move) stringify(board *Board) string {
	if board != nil {
		if (board.squares[move.start] / 2) == 1 && (move.end < A8+S || move.end > H1+N) {
			return string(FILE[move.start&7]) + string(RANK[move.start>>4]) + string(FILE[move.end&7]) + string(RANK[move.end>>4]) + "q"
		}
	}
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
	board.pieceCount[oldpiece] -= 1
	board.pieceCount[newpiece] += 1
}

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

func (board *Board) PawnDefends(sq int, attacker int8) bool {
	for _, dir := range []int{W - ADVANCES[attacker], E - ADVANCES[attacker]} {
		if ((sq + dir) & 0x88) == 0 {
			if (board.squares[sq + dir] == 2 + attacker) {
				return true
			}
		}
	}
	return false
}

func (board *Board) Generate(capturesOnly bool) ([]Move, int) {
	attention := &e_table[board.sidetomove]

	moves := []Move{}

	advance := ADVANCES[board.sidetomove]
	
	mobility := 0

	var pawnIndexes [16]int8
	pawnCounter := 0

	for i, piece := range board.squares {
		if piece < 2 {
			continue
		}

		piecetype := piece / 2

		if piecetype == 1 {
			pawnIndexes[pawnCounter] = int8(i)
			pawnCounter++
		}

		if piece & 1 != board.sidetomove {
			continue
		}


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
			ray := (piecetype / 3) == 1
			pattern := patterns[piecetype]
			
			for _, dir := range pattern {
				for end := i + dir; (end & 0x88) == 0; end += dir {
					victim := board.squares[end]

					if board.PawnDefends(end, 1-board.sidetomove) {
						mobility += decode(e_restricted[piecetype], board.phase)
					}
					
					if victim != 0 {
						if victim&1 != piece&1 {
							mobValue += attention[end]
							mobility += decode(e_attacks[piecetype * 7 + (victim/2)], board.phase)
							moves = append(moves, Move{int8(i), int8(end)})
						}
					} else {
						mobValue += attention[end]
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
		mobility += (decode(e_mobility[piecetype],board.phase) * mobValue) / e_divider
	}

	if board.enpassant != 0 {
		for _, enpassantStart := range []int{board.enpassant - advance + W, board.enpassant - advance + E} {
			if board.squares[enpassantStart] == 2 + board.sidetomove {
				moves = append(moves, Move{int8(enpassantStart), int8(board.enpassant)})
			}
		}
	}


	kingIndex := board.kings[board.sidetomove]
	board.inCheck = board.attacked(kingIndex, 1-board.sidetomove)
	if !capturesOnly && (kingIndex == E8 || kingIndex == E1) && !board.inCheck {
		if board.squares[kingIndex+E+E+E+CASTLE] == 1 && board.squares[kingIndex+E+E] == 0 && board.squares[kingIndex+E] == 0 {
			moves = append(moves, Move{int8(kingIndex), int8(kingIndex+E+E)})
		} 
		if board.squares[kingIndex+W+W+W+W+CASTLE] == 1 && board.squares[kingIndex+W+W+W] == 0 && board.squares[kingIndex+W+W] == 0 && board.squares[kingIndex+W] == 0 {
			moves = append(moves, Move{int8(kingIndex), int8(kingIndex+W+W)})
		}
	}


	score := 0

	for piecetype := range 7 {
		score += e_material[piecetype] * int((board.pieceCount[piecetype * 2 + 1] - board.pieceCount[piecetype * 2]))
	}


	whiterear := [10]int{0,0,0,0,0,0,0,0,0,0,}
	blackrear := [10]int{7,7,7,7,7,7,7,7,7,7,}

	for _, sq := range pawnIndexes {
		piece := board.squares[sq]
		if piece / 2 == 1 {
			pawnfile := (sq & 7) + 1
			pawnrank := int(sq >> 4)

			if (piece & 1) == 1 {
				whiterear[pawnfile] = max(whiterear[pawnfile], pawnrank)
			} else {
				blackrear[pawnfile] = min(blackrear[pawnfile], pawnrank)
			}
		}
	}


	wkingfile := (board.kings[1]&7) + 1
	bkingfile := (board.kings[0]&7) + 1
	//wkingrank := board.kings[1] >> 4
	//bkingrank := board.kings[0] >> 4

	for i := -1; i <= 1; i++ {
		if whiterear[wkingfile + i] != 0 {
			score += e_shield[wkingfile + i]
		}
		if blackrear[bkingfile + i] != 7 {
			score -= e_shield[bkingfile + i]
		}
	}


	npawns := [2]int{0,0}
	
	var whitepasser [10]int
	var blackpasser [10]int

	for _, sq := range pawnIndexes {
		piece := board.squares[sq]
		if piece / 2 == 1 {
			npawns[piece & 1] += 1
			pfile := int((sq & 7) + 1)
			prank := int(sq >> 4)

			semiopen := false
			
			if piece & 1 == 1 {
				semiopen = (blackrear[pfile] == 7)
				if blackrear[pfile - 1] >= prank && blackrear[pfile] >= prank && blackrear[pfile + 1] >= prank {
					whitepasser[pfile] = max(whitepasser[pfile], 7 - prank)
				}
			} else {
				semiopen = (whiterear[pfile] == 0)
				if whiterear[pfile - 1] <= prank && whiterear[pfile] <= prank && whiterear[pfile + 1] <= prank {
					blackpasser[pfile] = max(blackpasser[pfile], prank)
				}
			}

			if piece & 1 == board.sidetomove {
				
				if piece == board.squares[sq+W] || piece == board.squares[sq+E] {
					if semiopen {
						mobility += (decode(e_phalanxOpen, board.phase) * attention[sq]) / e_divider 
					} else {
						mobility += (decode(e_phalanxClosed, board.phase) * attention[sq]) / e_divider
					}
				}

				if board.PawnDefends(int(sq), board.sidetomove) {
					if semiopen {
						mobility += (decode(e_chainOpen, board.phase) * attention[sq]) / e_divider
					} else {
						mobility += (decode(e_chainClosed, board.phase) * attention[sq]) / e_divider 
					}
				}
			}
		}
	}

	for file := range 10 {
		if whitepasser[file] != 0 {
			score += e_passerRank[whitepasser[file]]
			score += e_passerFile[max(file - bkingfile, bkingfile - file)]
		}
		if blackpasser[file] != 0 {
			score -= e_passerRank[blackpasser[file]]
			score -= e_passerRank[max(file - wkingfile, wkingfile - file)]
		}
	}

	if board.pieceCount[6] == 2 {
		score -= e_bishopPair
	}

	if board.pieceCount[7] == 2 {
		score += e_bishopPair
	}

	if board.sidetomove == 1 {
		score += e_tempo
	} else {
		score -= e_tempo
	}

	score = decode(score, board.phase) 
	board.mobilities[board.sidetomove] = mobility
	score += board.mobilities[1] - board.mobilities[0]

	leadingpawns := board.pieceCount[2]
	if score >= 0 {
		leadingpawns = board.pieceCount[3]
	}
	
	score = (score * e_risk[leadingpawns]) / e_divider

	if board.sidetomove == 0 {
		return moves, -score
	}
	return moves, score
}

func (board *Board) attacked(start int, attacker int8) bool {
	if board.PawnDefends(start, attacker) {
		return true
	}

	for i, dir := range []int{N,S,E,W,N+W,N+E,S+E,S+W} {
		for sq := start + dir; (0x88 & sq) == 0; sq += dir {
			piece := board.squares[sq]
			if piece != 0 {
				if (i < 4 && (piece == 10 + attacker  || piece == 8 + attacker)) || (i >= 4 && (piece == 10 + attacker || piece == 6 + attacker)) {
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

		if (i >= 8 && (piece == 4 + attacker)) || (i < 8 && (piece == 12 + attacker)) {
			return true
		}
	}

	return false
}

func (board *Board) Apply(move Move) *Board {
	currentBoard := *board
	copyBoard := currentBoard

	newEP := 0
	
	if move.start != move.end {
		movingPiece := copyBoard.squares[move.start]
		copyBoard.Edit(int(move.start), 0)
		copyBoard.Edit(int(move.end), movingPiece)

		advance := ADVANCES[copyBoard.sidetomove]

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
			if copyBoard.sidetomove == 1 {
				copyBoard.Edit(A1 + CASTLE, 0)
				copyBoard.Edit(H1 + CASTLE, 0)

			} else {
				copyBoard.Edit(A8 + CASTLE, 0)
				copyBoard.Edit(H8 + CASTLE, 0)
			}
		}


		if copyBoard.squares[int(move.start)+CASTLE] != 0 {
			copyBoard.Edit(int(move.start)+CASTLE, 0)
		}
		if copyBoard.squares[int(move.end)+CASTLE] != 0 {
			copyBoard.Edit(int(move.end)+CASTLE, 0)
		}
	}

	if copyBoard.attacked(copyBoard.kings[copyBoard.sidetomove], 1 - copyBoard.sidetomove) {
		return nil
	}

	copyBoard.sidetomove = 1 - copyBoard.sidetomove
	copyBoard.enpassant = newEP
	copyBoard.inCheck = false
	copyBoard.ply += 1

	

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

    movelist, _ := perftboard.Generate(false)
    nodes := 0

    for i, move := range movelist {
        nextBoard := perftboard.Apply(move)
        if (nextBoard != nil) { 
        	subnodes := perft(nextBoard, depth - 1, maxdepth) 
        	nodes += subnodes 
        	if depth == maxdepth {
        		fmt.Println(i, move.stringify(nextBoard), subnodes)
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

type entry struct {
	key uint64
	move Move
	score int16
	depth int8
	bound int8
}

const hashsize = 16777216
var table [hashsize]entry

var history [14][128]int

func alphabeta(board *Board, alpha, beta, depth int, nullallowed bool) int {
	pv := beta - alpha != 1
	nodes += 1
	bestScore := -9999


	hash := board.Hash()
	tt := &table[hash % hashsize]

	if board.ply != 0 && depth > 0 {
		for _, rep := range repetition {
			if rep == hash {
				return 0
			}
		}
	}

	if !pv && tt.key == hash && (int(tt.depth) >= depth || 0 >= depth) {
		if (tt.bound == 1 && int(tt.score) <= alpha) {return int(tt.score)}
		if (tt.bound == -1 && int(tt.score) >= beta) {return int(tt.score)}
		if (tt.bound == 0) {return int(tt.score)}
	}

	// Weird Aspiration Window
	if pv && tt.key == hash && alpha == -10000 && beta == 10000 && depth >= 8 {
		cacheScore := int(tt.score)
		// Could try depth based width
		aspirationScore := alphabeta(board, cacheScore - 20, cacheScore + 20, depth, nullallowed)
		if cacheScore - 20 < aspirationScore && aspirationScore < cacheScore + 20 {
			return aspirationScore
		}

	}
     
	moves, staticEval := board.Generate(depth <= 0)
	// standpat
	if (depth <= 0) {
		bestScore = staticEval
		if bestScore > alpha {
			alpha = bestScore
		}
		if bestScore >= beta {
			return bestScore
		}
	}


	if depth > 0 && !pv && board.phase > 4 && !board.inCheck {
		// Reverse futility pruning RFP
		if (staticEval - ((depth * 50) + (5 * depth * depth)) > beta) { 
			return staticEval
		}

		// Null move pruning NMP
		if staticEval >= beta && nullallowed && depth >= 3 {
			nmScore := -alphabeta(board.Apply(Move{9,9}), -beta, -alpha, ((depth - 4) - (depth / 5)), false)
			if nmScore >= beta {
				return beta
			}
		}
	}


	// Prunings should only happen above this
	repetition = append(repetition, hash)


	priorities := make([]int, len(moves), len(moves))
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

	quietsLeft := (depth * depth) - depth + 4
	
	// Futility pruning
	if (depth <= 5 && staticEval + depth * 100 < alpha) {
		quietsLeft = 1
	}


	legals := 0
	var bestMove Move
	var boundtype int8 = 1

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

		if depth <= 0 && !pv && !board.inCheck { // Delta pruning
			if staticEval + decode(e_material[board.squares[nextMove.end]/2], board.phase) + 75 < alpha {
				break
			}
		}

		nextBoard := board.Apply(nextMove)
		if nextBoard == nil {
			continue
		}

		legals += 1

		var score int

		if ((legals == 1 || depth <= 0)){
			score = -alphabeta(nextBoard, -beta, -alpha, depth - 1, true)
		} else {
			// TODO: improve LMR
			reduction := ((depth+legals)/16)
			if board.squares[nextMove.end] == 0 {
				reduction += max(0,-max(-2, history[board.squares[nextMove.start]][nextMove.end] / 64))
			}
			
			score = -alphabeta(nextBoard, -alpha-1, -alpha, depth - 1 - reduction, true)
			if score > alpha && reduction > 0 {
				score = -alphabeta(nextBoard, -alpha-1, -alpha, depth - 1, true)
			}

			if score > alpha && pv {
				score = -alphabeta(nextBoard, -beta, -alpha, depth - 1, true)
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
			boundtype = -1

			if (board.squares[nextMove.end] == 0) {
				bonus := depth * depth
				if staticEval < alpha {
					bonus = (depth + 1) * (depth + 1)
				}

				hh := &history[board.squares[nextMove.start]][nextMove.end]

				*hh += bonus - ((bonus * (*hh)) / MAX_HISTORY)

				for m:=0;m<i;m++ {
					if board.squares[moves[m].end] == 0 {
						hhm := &history[board.squares[moves[m].start]][moves[m].end]
						*hhm -= (bonus + ((bonus * (*hhm)) / MAX_HISTORY))
					}
				}
			}

			break
		}
			
		if !pv && !board.inCheck && board.squares[nextMove.end] == 0 && legals != 1 {
			quietsLeft -= 1
			if quietsLeft == 0 {
				break
			}
		}
	}

	repetition = repetition[:len(repetition)-1]

	if legals == 0 && depth > 0 {
		if !board.inCheck {
			return 0
		}
	}

	if bestScore < -9000 {
		bestScore += 1 // Mate distance/delay adjustment
	}

	if bestMove.start != bestMove.end {
		table[hash % hashsize] = entry{hash, bestMove, int16(bestScore), int8(depth), boundtype}
	}

	return bestScore
}

var uciBoard Board
var repetition []uint64 = []uint64{}
   
func printpv() string {
	pvstr := ""
	board := &uciBoard
	for range 40 {
		if board == nil {
			break
		}
		m := table[board.Hash() % hashsize].move
		if board.Hash() != table[board.Hash() % hashsize].key {
			break
		}

		pvstr += m.stringify(board) + " "
		board = board.Apply(m)
	}
	return strings.TrimSpace(pvstr)
}

func parseuci(line string) bool {
	args := strings.Fields(line)

	if len(args) == 0 {
		return false
	}

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
		}
		uciBoard.ply = 0
	case "go":
		timeAlloc := 1000
		if len(findAfter("movetime", args)) != 0 {
			timeAlloc, _ = strconv.Atoi(findAfter("movetime", args)[0])
		}

		if len(findAfter("wtime", args)) != 0 && uciBoard.sidetomove == 1 {
			timeAlloc, _ = strconv.Atoi(findAfter("wtime", args)[0])
			timeAlloc /= 10
		}

		if len(findAfter("btime", args)) != 0 && uciBoard.sidetomove == 0 {
			timeAlloc, _ = strconv.Atoi(findAfter("btime", args)[0])
			timeAlloc /= 10
		}

		nodes = 0
		start := time.Now().UnixMilli()
		chosenMove := table[uciBoard.Hash() % hashsize].move.stringify(&uciBoard)
		streak := 0
		for depth := 1; depth <= 100; depth++ {
			fmt.Println("info score cp", alphabeta(&uciBoard, -10000, 10000, depth, true), "depth", depth, "time", time.Now().UnixMilli() - start, "nodes", nodes, "pv", printpv())

			if chosenMove == table[uciBoard.Hash() % hashsize].move.stringify(&uciBoard) {
				streak += 1
			} else {
				streak = 0
			}
			chosenMove = table[uciBoard.Hash() % hashsize].move.stringify(&uciBoard)

			if time.Now().UnixMilli() - start > int64(float64(timeAlloc) * math.Pow(0.9, float64(streak))) {
				break
			}
		}

		fmt.Println("bestmove", chosenMove)
	
	case "eval":
		uciBoard.sidetomove = 1 - uciBoard.sidetomove
		uciBoard.Generate(true)
		uciBoard.sidetomove = 1 - uciBoard.sidetomove
		_, e := uciBoard.Generate(true)

		fmt.Println("eval", e)

	/*
	case "null":
		uciBoard = *uciBoard.Apply(nullmove)
	*/

	case "quit":
		return true
	}
	return false
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func main() {
	flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
		if err != nil {
            fmt.Println(err)//log.Fatal("could not create CPU profile: ", err)
        }
        defer f.Close() // error handling omitted for example
        if err := pprof.StartCPUProfile(f); err != nil {
            fmt.Println(err) //log.Fatal("could not start CPU profile: ", err)
        }
        defer pprof.StopCPUProfile()
    }

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

	for {
		line, _ := reader.ReadString('\n')
		line = strings.Replace(line, "startpos", "fen rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 1)
		if parseuci(line) {
			break
		}
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
//         - opposite king passer bonus

// Still need to figure out how to reward the queen moreso than other pieces for attacking king ring


// Ben finegolds middle name is philip
// Should make a stream where people vote on best move


// TODO: squeeze more elo by optimizing pruning/reductions