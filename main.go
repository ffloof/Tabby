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

func BOOL(b bool) int { // golang for reasons unknown to me has no native way to convert a boolean to an integer
	if b {
		return 1
	}
	return 0
}

// 0.23725980257695797
var e_material = []int{T(0,0), T(39,82), T(291,298), T(317,321), T(401,618), T(853,1159), T(0,0), }
var e_passerRank = []int{T(0,0), T(3,0), T(-4,-12), T(-9,9), T(-1,36), T(2,101), T(-8,176), T(0,0), }
var e_shield = []int{T(0,0), T(17,27), T(53,17), T(2,26), T(20,10), T(31,5), T(28,9), T(8,12), T(69,-4), T(0,0), }
var e_tempo int = T(10,10) //{T(31,30), }
var e_passerFile = []int{T(0,0), T(19,8), T(4,7), T(-17,26), T(-18,26), T(-34,27), T(-9,6), T(-75,6), }
var e_restricted = []int{T(0,0), T(0,0), T(-6,-4), T(-5,0), T(-5,-1), T(-5,1), T(-15,5), }

var e_attacks = []int{T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(0,0), T(48,9), T(60,28), T(68,0), T(59,9), T(96,47), T(0,0), T(-7,12), T(0,0), T(20,45), T(37,39), T(27,15), T(100,1), T(0,0), T(-3,13), T(8,20), T(0,0), T(23,17), T(41,44), T(47,72), T(0,0), T(-15,12), T(-3,12), T(18,12), T(0,0), T(65,-8), T(194,-8), T(0,0), T(-1,6), T(-7,12), T(-1,37), T(2,6), T(0,0), T(62,115), T(0,0), T(29,30), T(6,17), T(-18,24), T(-126,48), T(-358,-92), T(0,0), }

var e_bishopPair int = T(18,39)

// Note mobility only counts current square for pawns
var e_mobility = []int{T(0,0), T(28,1), T(20,15), T(13,11), T(14,5), T(5,16), T(-24,22) }
var e_phalanxOpen int = T(6,11)
var e_phalanxClosed int = T(5,0)
var e_chainOpen int = T(18,21)
var e_chainClosed int = T(9,7)
var e_isolatedOpen int = T(-11,-3)
var e_isolatedClosed int = T(-1,-5)

var e_unstoppableVertical int = T(-273,117)
var e_protectedpasser int = T(52,-20)
var e_unstoppableHorizontal int = T(9,71)
var e_shield2 = []int{T(0,0), T(-31,25), T(15,13), T(-26,19), T(18,7), T(23,-3), T(11,0), T(-8,18), T(39,13), T(0,0), }

func flip(arr1, arr2 *[128]int, xor int){
	for i := range(len(arr1)) {
		arr2[i] = arr1[i^xor]
	}
}

var evals [256]int

var e_risk = []int{ 327, 534, 779, 932, 1052, 1044, 1054, 1062, 1000 }
var e_table = [2][128]int {
	{},
{-80, -129,  283,  161,  452,  560,  485,  785, 0,0,0,0, 0,0,0,0,
 425,  618,  505,  346,  524,  541,  681,  440, 0,0,0,0, 0,0,0,0,
 454,  520,  508,  373,  538,  832,  591,  296, 0,0,0,0, 0,0,0,0,
 369,  587,  395,  709,  683,  467,  535,  538, 0,0,0,0, 0,0,0,0,
 219,  249,  537,  720,  598,  363,  354,   97, 0,0,0,0, 0,0,0,0,
  24,  312,  423,  357,  488,  404,  793,  157, 0,0,0,0, 0,0,0,0,
  81,  406,  417,  497,  436,  635,  766,  151, 0,0,0,0, 0,0,0,0,
 -11,  146,  296,  346,  367,  -39,  131,   78, 0,0,0,0, 0,0,0,0,},
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

	board.sidetomove = int8(BOOL(fenparts[1] == "w"))

	board.Edit(H8+CASTLE, int8(BOOL(strings.Index(fenparts[2], "k") > -1)))
	board.Edit(A8+CASTLE, int8(BOOL(strings.Index(fenparts[2], "q") > -1)))
	board.Edit(H1+CASTLE, int8(BOOL(strings.Index(fenparts[2], "K") > -1)))
	board.Edit(A1+CASTLE, int8(BOOL(strings.Index(fenparts[2], "Q") > -1)))

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
	pawnIndexes := make([]int8, 0, 16)

	for i, piece := range board.squares {
		if piece < 2 {
			continue
		}

		piecetype := piece / 2

		if piecetype == 1 {
			pawnIndexes = append(pawnIndexes,int8(i))
		}

		if piece & 1 != board.sidetomove {
			continue
		}

		mobValue := attention[i] * 2
		if piecetype == 1 {
			if !capturesOnly && board.squares[i+advance] == 0 {
				moves = append(moves, Move{int8(i), int8(i + advance)})
				if ((board.sidetomove == 1 && A1+N <= i) || (board.sidetomove == 0 && i <= H8+S)) && board.squares[i+advance+advance] == 0 {
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

	// TODO: we could save a lot of time if we moved the incheck inside of the capturesOnly clause since we dont yet use it in eval, and its worthless in qsearch
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
		pawnfile := (sq & 7) + 1
		pawnrank := int(sq >> 4)

		if (piece & 1) == 1 {
			whiterear[pawnfile] = max(whiterear[pawnfile], pawnrank)
		} else {
			blackrear[pawnfile] = min(blackrear[pawnfile], pawnrank)
		}
	}


	wkingfile := (board.kings[1]&7) + 1
	bkingfile := (board.kings[0]&7) + 1
	wkingrank := board.kings[1] >> 4
	bkingrank := board.kings[0] >> 4

	for i := -1; i <= 1; i++ {
		if whiterear[wkingfile + i] != 0 {
			score += e_shield[wkingfile + i]
		} else if blackrear[wkingfile + i] != 7 {
			score += e_shield2[wkingfile + i]
		}

		if blackrear[bkingfile + i] != 7 {
			score -= e_shield[bkingfile + i]
		} else if whiterear[bkingfile + i] != 0 {
			score -= e_shield2[bkingfile + i]
		}
	}


	npawns := [2]int{0,0}
	
	var whitepasser [10]int
	var blackpasser [10]int

	// Weak pawn evaluation
	for _, sq := range pawnIndexes {
		piece := board.squares[sq]
		npawns[piece & 1] += 1
		pfile := int((sq & 7) + 1)
		prank := int(sq >> 4)

		semiopen := false
		
		if piece & 1 == 1 {
			semiopen = (blackrear[pfile] == 7)
			if blackrear[pfile - 1] >= prank && blackrear[pfile] >= prank && blackrear[pfile + 1] >= prank {
				whitepasser[pfile] = max(whitepasser[pfile], 7 - prank)
			}

			if piece == board.squares[sq+W] || piece == board.squares[sq+E] {
				if semiopen {
					score += e_phalanxOpen
				} else {
					score += e_phalanxClosed
				}
			}

			if board.PawnDefends(int(sq), 1) {
				if semiopen {
					score += e_chainOpen
				} else {
					score += e_chainClosed
				}
			}

			if whiterear[pfile-1] == 0 && whiterear[pfile+1] == 0 {
				if semiopen{
					score += e_isolatedOpen
				} else {
					score += e_isolatedClosed
				}
			}

		} else {
			semiopen = (whiterear[pfile] == 0)
			if whiterear[pfile - 1] <= prank && whiterear[pfile] <= prank && whiterear[pfile + 1] <= prank {
				blackpasser[pfile] = max(blackpasser[pfile], prank)
			}

			if piece == board.squares[sq+W] || piece == board.squares[sq+E] {
				if semiopen {
					score -= e_phalanxOpen
				} else {
					score -= e_phalanxClosed
				}
			}

			if board.PawnDefends(int(sq), 0) {
				if semiopen {
					score -= e_chainOpen
				} else {
					score -= e_chainClosed
				}
			}

			if blackrear[pfile-1] == 7 && blackrear[pfile+1] == 7 {
				if semiopen{
					score -= e_isolatedOpen
				} else {
					score -= e_isolatedClosed
				}
			}
		}

		/*
		if piece & 1 == board.sidetomove {
			
			if piece == board.squares[sq+W] || piece == board.squares[sq+E] {
				if semiopen {
					mobility += decode(e_phalanxOpen, board.phase) 
				} else {
					mobility += decode(e_phalanxClosed, board.phase)
				}
			}

			if board.PawnDefends(int(sq), board.sidetomove) {
				if semiopen {
					mobility += decode(e_chainOpen, board.phase)
				} else {
					mobility += decode(e_chainClosed, board.phase) 
				}
			}
		}*/
	}


	// Passer evaluation
	for file := range 10 {

		if whitepasser[file] != 0 {
			sq := ((7-whitepasser[file])*16) + (file - 1)
			score += e_passerRank[whitepasser[file]]
			score += e_passerFile[max(file - bkingfile, bkingfile - file)]
			
			if board.PawnDefends(sq, 1) {
				score += e_protectedpasser
				
			}
			
			if 7-whitepasser[file] < bkingrank - (1-int(board.sidetomove)){
                score += e_unstoppableVertical
                
			}
            if (7-whitepasser[file] < max(bkingfile-file,file-bkingfile)-(1-int(board.sidetomove))){
                score += e_unstoppableHorizontal
            }


		}
		if blackpasser[file] != 0 {
			sq := (blackpasser[file]*16) + (file - 1)
			score -= e_passerRank[blackpasser[file]]
			score -= e_passerFile[max(file - wkingfile, wkingfile - file)]
			if board.PawnDefends(sq, 0) {
				score -= e_protectedpasser
			}

			if blackpasser[file] > wkingrank + int(board.sidetomove) {
                score -= e_unstoppableVertical
			}
            if (7-blackpasser[file] < max(wkingfile-file,file-wkingfile)-int(board.sidetomove)){
                score -= e_unstoppableHorizontal
            }
		}
	}

	// Bishop pair evaluation
	if board.pieceCount[6] == 2 {
		score -= e_bishopPair
	}

	if board.pieceCount[7] == 2 {
		score += e_bishopPair
	}

	// Tempo evaluation
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

var history [2][14][128]int

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
	evals[board.ply] = staticEval

	improving := board.ply > 1 && staticEval > evals[board.ply - 2];

	if tt.key == hash {
		staticEval = int(tt.score)
	}

	// standpat
	if (depth <= 0) {
		bestScore = staticEval
		alpha = max(alpha, bestScore)
		if bestScore >= beta {
			return bestScore
		}
	}


	if depth > 0 && !pv && !board.inCheck {
		// Reverse futility pruning RFP
		if (staticEval - ((depth * 50) + (5 * depth * depth)) > beta) { 
			return staticEval
		}

		// Null move pruning NMP
		if staticEval >= beta && nullallowed && depth >= 3 && board.phase > 4 {
			if -alphabeta(board.Apply(Move{9,9}), -beta, -alpha, ((depth - 4) - (depth / 5)) - ((staticEval - beta)/200), false) >= beta {
				return beta
			}
		}
	}


	// Prunings that return should only happen above this
	repetition = append(repetition, hash)

	priorities := make([]int, len(moves), len(moves))
	for i, move := range moves {
		priorities[i] = (int(board.squares[move.end]) * 1_000_000) + history[BOOL(board.squares[move.end] == 0)][board.squares[move.start]][move.end]
		
		if tt.move.end == move.end && tt.move.start == move.start {
			priorities[i] = 1_000_000_000
		}
	}

	quietsLeft := ((depth * depth + 1) >> BOOL(!improving)) + 1
	
	// Futility pruning
	if (depth <= 5 && staticEval + depth * 100 < alpha) {
		quietsLeft = 0
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

		if !pv && !board.inCheck && board.squares[nextMove.end] == 0 && legals != 0 {
			if quietsLeft <= 0 {
				break
			}
			quietsLeft -= 1
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
			reduction := ((depth+legals)/16)
			reduction += max(-2,-max(-2, history[BOOL(board.squares[nextMove.end] == 0)][board.squares[nextMove.start]][nextMove.end] / 64))
			// TODO: test if we should just not reduce captures
			reduction = max(reduction, 0)

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

			bonus := depth * depth
			hh := &history[BOOL(board.squares[nextMove.end] == 0)][board.squares[nextMove.start]][nextMove.end]
			*hh += bonus - ((bonus * (*hh)) / MAX_HISTORY)

			for m:=0;m<i;m++ {
				hhm := &history[BOOL(board.squares[moves[m].end] == 0)][board.squares[moves[m].start]][moves[m].end]
				*hhm -= (bonus + ((bonus * (*hhm)) / MAX_HISTORY))
			}
			break
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

	if bestMove.start != bestMove.end { // Many nodes in Qsearch either have no captures, or (in check) no legal captures, not worth storing them in tt since they are cheap and plentiful
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
	case "ucinewgame":
		history = [2][14][128]int{}
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




// Ben finegolds middle name is philip
// Should make a stream where people vote on best move