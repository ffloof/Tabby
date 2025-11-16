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

/*
import ("flag"
	"runtime/pprof"
)*/

const N, S, E, W = -16, 16, 1, -1
const A8, H8, A1, H1 = 0, 7, 112, 119
const E8, E1 = 4, 116
const CASTLE = E * 8
const PIECE = " .pPnNbBrRqQkK"
const RANK = "87654321"
const FILE = "abcdefgh"
var ADVANCES = [...]int{S,N}

func BOOL(b bool) int { // golang for reasons unknown to me has no native way to convert a boolean to an integer
	if b {
		return 1
	}
	return 0
}

var e_contempt = 0
var e_material = []int{0, 125, 436, 482, 900, 1824, 0, }
var e_tempo = 16
var e_passerRank = []int{0, 5, 16, 42, 90, 170, 250, 0, }
var e_phalanx = []int{7,32}
var e_chain = []int{22,51}
var e_baseMobility = []int{0, 0, 12, 9, 5, 0, 2, }
var e_passerKingDistance = []int{-79, -37, -22, -20, 0, 8, 25, 14, 66, }
var e_passerSupportDistance = []int{38, 43, 29, 4, 0, -3, 5, -17, -3,}
var e_bishopPair = 65
var e_egKingFile = []int{-90, -31, -22, 5, 0, 0, -42, -97, }
var e_egKingRank = []int{55, 159, 165, 167, 137, 90, 65, 0, }

var e_mobility = []int{0, 67, 34, 20, 23, 17, -20, }

var e_altmaterial = []int{0, -14, 75, 93, 46, 112, 0, }
var e_shield = []int{0, 63, 66, 33, 11, 25, 13, 42, 35, 0, }
var e_shieldbase = []int{0, -16, 6, 17, -3, 4, 27, 10, -2, 0, }
var e_mgKingFile = []int{19, -19, -12, -31, 0, -17, 33, 86, }
var e_mgKingRank = []int{-6, -20, 0, -90, -103, -74, -34, 0, }
var e_pawnattacked = 94

var e_risk = []int{3423, 1799,  949,  549,  323,  155,    0,  124,  328}

var e_table = [128]int { 
  -233,  -782,   -94,   163,   219,   425,   718,  1858,   0,0,0,0, 0,0,0,0, 
   194,   466,   503,   380,   401,   309,  1181,   590,   0,0,0,0, 0,0,0,0, 
   192,   541,   465,   396,   792,  1066,  1365,   794,   0,0,0,0, 0,0,0,0, 
   156,   326,   357,   457,   576,   582,   451,   340,   0,0,0,0, 0,0,0,0, 
    45,   174,   237,   391,   434,   280,   200,    43,   0,0,0,0, 0,0,0,0, 
   -86,    69,    49,   219,   240,   208,   324,   -58,   0,0,0,0, 0,0,0,0, 
  -125,    76,   184,    39,   134,   353,   248,   -84,   0,0,0,0, 0,0,0,0, 
     7,   -50,    30,   -43,    17,   208,  -116,   156,   0,0,0,0, 0,0,0,0, }

var weight_table = [2][4][128]int{}

const e_divider = 1000

var phaseWeights = [14]int{0, -19, 69, 107, 140, 463, 66, }

var Zobrist [16][128]uint64
var nodes int = 0
const MAX_HISTORY = 256
var evals [256]int

type Board struct {
	squares    [128]int8
	pieceCount [16]int8
	kings      [2]int
	enpassant  int
	zobrist    uint64
	mobilities [2][2]int
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
	board.zobrist ^= Zobrist[oldpiece][index]
	board.zobrist ^= Zobrist[newpiece][index]
	board.pieceCount[oldpiece] -= 1
	board.pieceCount[newpiece] += 1
	board.squares[index] = newpiece
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
	attention := &weight_table[board.sidetomove][(board.kings[1-board.sidetomove]%8)/2]

	moves := []Move{}

	advance := ADVANCES[board.sidetomove]
	baseMobility, dynamicMobility := 0, 0
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

		mobValue := attention[i]
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
						dynamicMobility += e_pawnattacked * BOOL(victim > 3)
					}
				}
			}
		} else {
			ray := (piecetype / 3) == 1
			pattern := patterns[piecetype]
			
			for _, dir := range pattern {
				for end := i + dir; (end & 0x88) == 0; end += dir {
					victim := board.squares[end]
					baseMobility += e_baseMobility[piecetype]
					
					if victim == 0 || victim&1 != piece&1 {
						if !board.PawnDefends(end, 1-board.sidetomove) {
							mobValue += attention[end]
						}

						if !capturesOnly || victim != 0 {
							moves = append(moves, Move{int8(i), int8(end)})
						}
					}

					if victim != 0 || !ray {
						break
					}
				}
			}
		}
		dynamicMobility += (e_mobility[piecetype] * mobValue) / e_divider
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

	score, dynamic_score, dynamics_weight := 0,0,0
	dynamic_score += (e_contempt * BOOL((board.ply - int(board.sidetomove)) % 2 == 1)) - (e_contempt * BOOL((board.ply - int(board.sidetomove)) % 2 == 0))  


	for piecetype := range 7 {
		score += e_material[piecetype] * int((board.pieceCount[piecetype * 2 + 1] - board.pieceCount[piecetype * 2]))
		dynamic_score += e_altmaterial[piecetype] * int((board.pieceCount[piecetype * 2 + 1] - board.pieceCount[piecetype * 2]))
		dynamics_weight += phaseWeights[piecetype]* int((board.pieceCount[piecetype * 2 + 1] + board.pieceCount[piecetype * 2]))
	}

	score += e_bishopPair * (BOOL(board.pieceCount[7] == 2) - BOOL(board.pieceCount[6] == 2))

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

	score += e_egKingFile[board.kings[1]&7] + e_egKingRank[board.kings[1]>>4]
	score -= e_egKingFile[board.kings[0]&7] + e_egKingRank[7-(board.kings[0]>>4)]
	dynamic_score += e_mgKingFile[board.kings[1]&7] + e_mgKingRank[board.kings[1]>>4]
	dynamic_score -= e_mgKingFile[board.kings[0]&7] + e_mgKingRank[7-(board.kings[0]>>4)]

	wkingfile := (board.kings[1]&7) + 1
	bkingfile := (board.kings[0]&7) + 1

	for i := -1; i <= 1; i++ {
		if whiterear[wkingfile + i] == 6 {
			dynamic_score += e_shieldbase[wkingfile + i]
		}
		if whiterear[wkingfile + i] != 0 {
			dynamic_score += e_shield[wkingfile + i]
		}

		if blackrear[bkingfile + i] == 1 {
			dynamic_score -= e_shieldbase[bkingfile + i]
		}
		if blackrear[bkingfile + i] != 7 {
			dynamic_score -= e_shield[bkingfile + i]
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

		if piece & 1 == 1 {
			semiopen := BOOL(blackrear[pfile] == 7) // TODO: can probably inline semi open
			if blackrear[pfile - 1] >= prank && blackrear[pfile] >= prank && blackrear[pfile + 1] >= prank {
				whitepasser[pfile] = max(whitepasser[pfile], 7 - prank)
			}

			if piece == board.squares[sq+W] || piece == board.squares[sq+E] {
				score += e_phalanx[semiopen]
			}

			if board.PawnDefends(int(sq), 1) {
				score += e_chain[semiopen]
			}
		} else {
			semiopen := BOOL(whiterear[pfile] == 0)
			if whiterear[pfile - 1] <= prank && whiterear[pfile] <= prank && whiterear[pfile + 1] <= prank {
				blackpasser[pfile] = max(blackpasser[pfile], prank)
			}

			if piece == board.squares[sq+W] || piece == board.squares[sq+E] {
				score -= e_phalanx[semiopen]
			}

			if board.PawnDefends(int(sq), 0) {
				score -= e_chain[semiopen]
			}
		}
	}

	// Passer evaluation
	for file := range 10 {
		if whitepasser[file] != 0 {
			score += e_passerRank[whitepasser[file]]
			score += e_passerKingDistance[max(8*BOOL(whitepasser[file]<(board.kings[0] >> 4)),max(file - bkingfile, bkingfile - file))]
			score += e_passerSupportDistance[max(8*BOOL(whitepasser[file]<(board.kings[1] >> 4)),max(file - wkingfile, wkingfile - file))]
		}
		if blackpasser[file] != 0 {
			score -= e_passerRank[blackpasser[file]]
			score -= e_passerKingDistance[max(8*BOOL(blackpasser[file]>(board.kings[1] >> 4)),max(file - wkingfile, wkingfile - file))]
			score -= e_passerSupportDistance[max(8*BOOL(blackpasser[file]>(board.kings[0] >> 4)),max(file - bkingfile, bkingfile - file))]
		}
	}

	board.mobilities[board.sidetomove][0] = baseMobility
	board.mobilities[board.sidetomove][1] = dynamicMobility
	
	score += board.mobilities[1][0] - board.mobilities[0][0]
	dynamic_score += board.mobilities[1][1] - board.mobilities[0][1]

	final_score := score + (dynamic_score * max(0,dynamics_weight)) / e_divider

	leadingpawns := board.pieceCount[2 + BOOL(final_score >= 0)]

	drawishness_weight := e_risk[leadingpawns]
	final_score = (final_score * e_divider) / (e_divider + dynamics_weight + drawishness_weight) 
	

	if board.sidetomove == 0 {
		final_score = -final_score
	}
	return moves, final_score + e_tempo // TODO: maybe we should properly address tempo in tuning
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
	} else if (depth > 3) {
		depth--
	}

	// standpat
	if (depth <= 0) {
		bestScore = staticEval
		alpha = max(alpha, bestScore)
		if bestScore >= beta {
			return bestScore
		}
	} else if board.inCheck {
		depth++
	}

	if depth > 0 && !pv && !board.inCheck {
		// Reverse futility pruning RFP
		if (staticEval - ((20 * depth * depth) + 50) > beta) { 
			return staticEval
		}

		// Null move pruning NMP
		if staticEval >= beta && nullallowed && depth >= 3 && (0 < board.pieceCount[4+board.sidetomove] + board.pieceCount[6+board.sidetomove] + board.pieceCount[8+board.sidetomove] + board.pieceCount[10+board.sidetomove]) {
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

	quietsLeft := ((depth * depth - 2 * depth + 4) >> BOOL(!improving)) + 1

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
			if staticEval + e_material[board.squares[nextMove.end]/2] + 25 < alpha {
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
			reduction := (depth+legals)/16
			reduction += -history[BOOL(board.squares[nextMove.end] == 0)][board.squares[nextMove.start]][nextMove.end] / 64  // default 64
			reduction = max(min(reduction,1+depth/3), 0) // 1 + depth/3 gained like 2 elo

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

	bestScore += BOOL(bestScore < -9000) // Mate distance/delay adjustment

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


/*
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
*/

func main() {
	/*
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
    }*/

	fmt.Println("info string Started")
	for i := range 15 {
		for j := range 128 {
			Zobrist[i+1][j] = rand.Uint64()
		}
	}

	for a := range e_table {
		for b := range 4 {
			weight_table[0][b][a] = (b * e_table[a^112] + (3-b) * e_table[a^112^7]) / 3
			weight_table[1][b][a] = (b * e_table[a]     + (3-b) * e_table[a^7])     / 3
		}
	}

	uciBoard = FromFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	reader := bufio.NewReader(os.Stdin)

	for {
		line, _ := reader.ReadString('\n')
		line = strings.Replace(line, "startpos", "fen rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 1)
		args := strings.Fields(line)

		if len(args) != 0 {
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
			case "position":
				uciBoard = FromFen(strings.Join(findAfter("fen", args)[0:4], " "))
				for _, movestr := range findAfter("moves", args) {
					uciMove := Move{int8(Parse(movestr[0:2])), int8(Parse(movestr[2:4]))}
					uciBoard = *(uciBoard.Apply(uciMove))
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
			case "quit":
				return
			}
		}
	}
}

// Should make a stream where people vote on best move4