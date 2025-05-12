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
19 0.24038275860488212

Linear terms
{T(0,0), T(43,80), T(285,264), T(328,293), T(383,544), T(825,1019), T(0,0), }
{T(0,0), T(7,10), T(42,10), T(4,20), T(16,8), T(14,9), T(13,12), T(20,6), T(29,-3), T(0,0), }
{T(0,0), T(10,-25), T(1,-18), T(-3,6), T(5,40), T(-5,117), T(-20,202), T(0,0) }
{T(-11,-9), }
{T(-5,1), }
{T(0,0), T(14,6), T(3,16), T(-12,33), T(-17,49), T(-31,60), T(-13,54), T(-67,73), }
{T(0,0), T(0,0), T(-8,-6), T(-6,-1), T(-10,2), T(-6,2), T(-23,14), }
{T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(37,2), T(64,37), T(76,-9), T(46,6), T(122,24), T(0,0), T(-7,10), T(0,0), T(21,38), T(33,35), T(29,28), T(107,-2), T(0,0), T(-2,9), T(11,33), T(0,0), T(23,24), T(39,36), T(48,73), T(0,0), T(-15,13), T(-1,16), T(14,18), T(0,0), T(49,-19), T(173,-12), T(0,0), T(0,5), T(-6,9), T(-4,37), T(-4,0), T(0,0), T(55,101), T(0,0), T(34,28), T(10,22), T(-16,30), T(-118,44), T(-354,-79), T(0,0), }
{T(30,33), }

Mobility weights
{T(0,0), T(58,3), T(21,15), T(11,7), T(18,5), T(4,17), T(-12,14), }

Risk weights
[0.334 0.595 0.82  0.935 1.046 1.042 1.172 1.078 1.   ]

Board Weights
[[-0.064 -0.297  0.049  0.207  0.401  0.386  0.787  1.231]
 [ 0.425  0.703  0.589  0.369  0.616  0.366  0.887  0.591]
 [ 0.279  0.603  0.758  0.358  0.492  0.759  0.474  0.252]
 [ 0.324  0.538  0.36   0.59   0.604  0.368  0.514  0.457]
 [ 0.092  0.238  0.43   0.674  0.586  0.387  0.338  0.16 ]
 [ 0.033  0.303  0.43   0.403  0.591  0.36   0.864  0.212]
 [ 0.002  0.221  0.313  0.382  0.373  0.791  0.677  0.134]
 [ 0.523  0.214  0.285  0.191  0.298  0.672  0.024 -0.038]]
===
*/

var e_material = []int{T(0,0), T(43,80), T(285,264), T(328,293), T(383,544), T(825,1019), T(0,0), }
var e_shield = []int{T(0,0), T(7,10), T(42,10), T(4,20), T(16,8), T(14,9), T(13,12), T(20,6), T(29,-3), T(0,0), }
var e_passerRank = []int{T(0,0), T(10,-25), T(1,-18), T(-3,6), T(5,40), T(-5,117), T(-20,202), T(0,0), }
var e_isolated int = T(-11,-9)
var e_backwards int = T(-5,1)
var e_passerFile = []int{T(0,0), T(14,6), T(3,16), T(-12,33), T(-17,49), T(-31,60), T(-13,54), T(-67,73), }
var e_restricted = []int{T(0,0), T(0,0), T(-8,-6), T(-6,-1), T(-10,2), T(-6,2), T(-23,14), }

var e_attacks = []int{
	T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), 
	T(0,0), T(0,0), T(37,2), T(64,37), T(76,-9), T(46,6), T(122,24), 
	T(0,0), T(-7,10), T(0,0), T(21,38), T(33,35), T(29,28), T(107,-2), 
	T(0,0), T(-2,9), T(11,33), T(0,0), T(23,24), T(39,36), T(48,73), 
	T(0,0), T(-15,13), T(-1,16), T(14,18), T(0,0), T(49,-19), T(173,-12), 
	T(0,0), T(0,5), T(-6,9), T(-4,37), T(-4,0), T(0,0), T(55,101), 
	T(0,0), T(34,28), T(10,22), T(-16,30), T(-118,44), T(-354,-79), T(0,0), 
}

//var e_tempo int = T(30,33)
var e_tempo int = T(10,10)

// Note mobility only counts current square for pawns
var e_mobility = []int{T(0,0), T(58,3), T(21,15), T(11,7), T(18,5), T(4,17), T(-12,14), }

func flip(arr1, arr2 *[128]int, xor int){
	for i := range(len(arr1)) {
		arr2[i] = arr1[i^xor]
	}
}

var e_risk = []int{ 334, 595, 820, 935, 1046, 1042, 1172, 1078, 1000 }
var e_table = [2][128]int {
	{},
{ -64, -297,   49,  207,  401,  386,  787,  231,     0,0,0,0, 0,0,0,0,
  425,  703,  589,  369,  616,  366,  887,  591,     0,0,0,0, 0,0,0,0,
  279,  603,  758,  358,  492,  759,  474,  252,     0,0,0,0, 0,0,0,0,
  324,  538,  360,  590,  604,  368,  514,  457,     0,0,0,0, 0,0,0,0,
   92,  238,  430,  674,  586,  387,  338,  160,     0,0,0,0, 0,0,0,0,
   33,  303,  430,  403,  591,  360,  864,  212,     0,0,0,0, 0,0,0,0,
    2,  221,  313,  382,  373,  791,  677,  134,     0,0,0,0, 0,0,0,0,
  523,  214,  285,  191,  298,  672,  024,  -38,     0,0,0,0, 0,0,0,0,},
}

const e_divider = 1000

var phaseWeights = [14]int{0,0,0,0,1,1,1,1,2,2,4,4,0,0}

var Zobrist [16][128]uint64

var nodes int = 0
const MAX_HISTORY = 8192


type Board struct {
	squares    [128]int8
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
var nullmove Move = Move{9,9}

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
			
			for _, dir := range pattern {
				for end := i + dir; (end & 0x88) == 0; end += dir {
					victim := board.squares[end]

					if (end - advance + W) & 0x88 == 0 {
						if board.squares[end - advance + W] == (3-board.sidetomove) {
							mobility += decode(e_restricted[piecetype], board.phase)
						}
					}

					if (end - advance + E) & 0x88 == 0 {
						if board.squares[end - advance + E] == (3-board.sidetomove) {
							mobility += decode(e_restricted[piecetype], board.phase)
						}
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
		//fmt.Println(mobValue)
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
	board.inCheck = board.attacked(kingIndex, 1-board.sidetomove)
	if !capturesOnly && (kingIndex == E8 || kingIndex == E1) && !board.inCheck {
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
			//fmt.Println(move.end)

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

    movelist := perftboard.GenerateLegalMoves(false)
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

	for sq, piece := range board.squares {
		if piece / 2 == 1 {
			npawns[piece & 1] += 1
			pfile := (sq & 7) + 1
			prank := sq >> 4

			if piece & 1 == 1 {
				if whiterear[pfile-1] == 0 && whiterear[pfile+1] == 0 {
					score += e_isolated
				}
				if whiterear[pfile - 1] < prank && whiterear[pfile + 1] < prank {
					score += e_backwards
				}
				if blackrear[pfile - 1] >= prank && blackrear[pfile] >= prank && blackrear[pfile + 1] >= prank {
					whitepasser[pfile] = max(whitepasser[pfile], 7 - prank)
				}
			} else {
				if blackrear[pfile-1] == 7 && blackrear[pfile+1] == 7 {
					score -= e_isolated
				}
				if blackrear[pfile - 1] > prank && blackrear[pfile + 1] > prank {
					score -= e_backwards
				}
				if whiterear[pfile - 1] <= prank && whiterear[pfile] <= prank && whiterear[pfile + 1] <= prank {
					blackpasser[pfile] = max(blackpasser[pfile], prank)
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

	if board.sidetomove == 1 {
		score += e_tempo
	} else {
		score -= e_tempo
	}

	score = decode(score, board.phase) 
	score += board.mobilities[1] - board.mobilities[0]

	//fmt.Println(board.mobilities)

	leadingpawns := npawns[0]
	if score >= 0 {
		leadingpawns = npawns[1]
	}
	
	score = (score * e_risk[leadingpawns]) / e_divider

	if board.sidetomove == 0 {
		return -score
	}
	return score
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

func alphabeta(board *Board, alpha, beta, depth int, nullallowed bool) int {
	pv := beta - alpha != 1
	nodes += 1
	bestScore := -9999 + board.ply

	moves := board.GenerateLegalMoves(depth <= 0)
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


	hash := board.Hash()

	if board.ply != 0 {
		for _, rep := range repetition {
			if rep == hash {
				return 0
			}
		}
	}


	tt := table[hash % hashsize]

	if !pv && tt.key == hash && (tt.depth >= depth || 0 >= depth) {
		if (tt.bound == 1 && tt.score <= alpha) {return tt.score}
		if (tt.bound == -1 && tt.score >= beta) {return tt.score}
		if (tt.bound == 0) {return tt.score}
	}
	

	if depth > 0 && !pv && board.phase > 4 && !board.inCheck {
		// Reverse futility pruning RFP
		if (depth < 8 && staticEval - depth * 100 > beta) { 
			return staticEval
		}

		// Null move pruning NMP
		if staticEval >= beta && nullallowed && depth >= 3 {
			nmBoard := board.Apply(nullmove)
			if nmBoard != nil {
				nmScore := -alphabeta(nmBoard, -beta, -alpha, depth - 3 - depth / 6, false)
				if nmScore >= beta {
					return beta
				}
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

	quietsLeft := (depth * depth) - depth + 5
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



		nextBoard := board.Apply(nextMove)
		if nextBoard == nil {
			continue
		}

		legals += 1
		reduction := (depth+legals)/16

		var score int

		if ((legals == 1 || depth <= 0)){
			score = -alphabeta(nextBoard, -beta, -alpha, depth - 1, true)
		} else {
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
						*hhm -= (bonus - ((bonus * (*hhm)) / MAX_HISTORY))
					}
				}
			}

			break
		}

		quietsLeft = quietsLeft
			
		if !pv && board.squares[nextMove.end] == 0 {
			quietsLeft -= 1
			if quietsLeft == 0 {
				break
			}
		}
	}

	repetition = repetition[:len(repetition)-1]

	if legals == 0 && depth > 0 {
		if board.inCheck {
			return bestScore // Best score is mate score
		} else {
			return 0
		}
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
	for range 40 {
		if board == nil {
			break
		}
		m := table[board.Hash() % hashsize].move
		if board.Hash() != table[board.Hash() % hashsize].key {
			break
		}

		pvstr += m.stringify(board) + " "

		//fmt.Println(m.stringify(), table[board.Hash() % hashsize].depth)
		board = board.Apply(m)
	}
	return strings.TrimSpace(pvstr)
}

func parseuci(line string) {
	args := strings.Fields(line)

	if len(args) == 0 {
		return
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
			timeAlloc /= 30
		}

		if len(findAfter("btime", args)) != 0 && uciBoard.sidetomove == 0 {
			timeAlloc, _ = strconv.Atoi(findAfter("btime", args)[0])
			timeAlloc /= 30
		}

		nodes = 0
		start := time.Now().UnixMilli()
		for depth := 1; depth <= 100; depth++ {
			fmt.Println("info score cp", alphabeta(&uciBoard, -10000, 10000, depth, true), "depth", depth, "time", time.Now().UnixMilli() - start, "nodes", nodes, "pv", printpv())

			if int(time.Now().UnixMilli() - start) > timeAlloc {
				break
			}
		}

		fmt.Println("bestmove", table[uciBoard.Hash() % hashsize].move.stringify(&uciBoard))
	
	case "eval":
		uciBoard.GenerateLegalMoves(true)
		uciBoard.sidetomove = 1 - uciBoard.sidetomove
		uciBoard.GenerateLegalMoves(true)
		uciBoard.sidetomove = 1 - uciBoard.sidetomove

		fmt.Println("eval", eval(&uciBoard))

	case "null":
		uciBoard = *uciBoard.Apply(nullmove)


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
//         - opposite king passer bonus

// Still need to figure out how to reward the queen moreso than other pieces for attacking king ring


// Ben finegolds middle name is philip
// Should make a stream where people vote on best move

// Base Search      elo     W/D/L
// + RFP        ~ 120 elo 81/21/34
// + LMP        ~ 60 elo  53/37/32?
// + NMP        ~ 60 elo  56/36/32
// + LMR         ~ 50 elo  170/100/116
// + fixed tt    ~ 100 elo 40/23/18
// TODO: squeeze more elo by optimizing pruning/reductions