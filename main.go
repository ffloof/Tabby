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

const N, S, E, W = -10, 10, 1, -1
const A8, H8, A1, H1 = 21, 28, 91, 98
const E8, E1 = 25, 95
const PIECE = " .pPnNbBrRqQkK"
const RANK = "  87654321  "
const FILE = " abcdefgh "
var ADVANCES = [...]int{S,N}

func BOOL(b bool) int { // golang for reasons unknown to me has no native way to convert a boolean to an integer
	if b {
		return 1
	}
	return 0
}

var e_material = []int{0, 89, 378, 414, 662, 1340, 0, }
var e_tempo = 15
var e_passerRank = []int{0, 3, 1, 17, 47, 132, 213, 0, }
var e_phalanx = []int{6, 13, }

var e_egKingFile = []int{0, -54, -25, -9, 0, -10, -4, -40, -58, 0, } //B
var e_egKingRank = []int{14, 47, 59, 59, 64, 51, 31, 0, } //B

var e_mobility = []int{0, 51, 41, 34, 27, 15, -24, }

var e_altmaterial = []int{0, -44, 0, 0, 0, 0, 0, }
var e_shield = []int{0, 103, 101, 30, 23, 20, 12, 48, 25, 0, } //A
var e_shieldbase = []int{0, -34, 25, 18, 13, 16, 28, 47, 2, 0, } //A
var e_mgKingFile = []int{0, 17, 23, 40, 0, 73, 33, 113, 134, 0, } //B
var e_mgKingRank = []int{135, 102, 165, 24, -95, -85, -52, 0, } //B
var e_pawnattacked = 100

var e_risk = []int{3269,  956,  376,   -5,  -31,    0,    0,    0,   -9,} 
var phaseWeights = [14]int{0,  -1,   9,  35,  51, 168,  31, } //F 3 LOC

var e_table = [64]int { // 20+ LOC
 363, -657,  416,  506,  792,  900, 1068, 2194,
 225,  471,  802,  629, 1083,  881, 1855, 1154,
 556,  861,  893,  457, 1872, 1950, 2226,  772,
 725,  698,  478, 1054, 1031, 1068,  832, 1162,
 606,  347,  816, 1129, 1278,  785,  381,  452,
 -92,  448,  669,  717,  715,  697,  794,   93,
-258,  336,  595,  544,  485,  813,  599, -141,
  57,  306,  521,  418,  610,  813,  330, -742,}

var weight_table = [2][4][120]int{}

const e_divider = 1000
var e_contempt = 0

var Zobrist [16][120]uint64
var nodes int = 0
const MAX_HISTORY = 256
var evals [256]int

type Board struct {
	squares    [120]int8
	pieceCount [16]int8
	kings      [2]int
	enpassant  int
	zobrist    uint64
	mobilities [2]int
	castleRights [2][2]bool
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
			return string(FILE[move.start%10]) + string(RANK[move.start/10]) + string(FILE[move.end%10]) + string(RANK[move.end/10]) + "q"
		}
	}
	return string(FILE[move.start%10]) + string(RANK[move.start/10]) + string(FILE[move.end%10]) + string(RANK[move.end/10])
}

func Parse(sqstr string) int {
	return strings.Index(FILE, sqstr[0:1]) + strings.Index(RANK, sqstr[1:2]) * 10
}

func FromFen(fen string) Board {
	board := Board{}
	fenparts := strings.Split(fen, " ")

	// TODO: for morning I had this crazy idea, what if we just spam replaces to create the board via string and then convert it all in one loop
	// wicked
	nextStop := A8
	for i := range 120 {
		board.Edit(i,int8(BOOL(i % 10 == 0 || i % 10 == 9 || i < A8 || i > H1)))
		if i == nextStop && nextStop <= H1 {
			char := fenparts[0][0:1]
			fenparts[0] = fenparts[0][1:]
			piece := int8(strings.Index(PIECE, char))
			nextStop += 1

			if piece >= 0 {
				board.Edit(i, piece)
				if piece >= 12 {
					board.kings[piece & 1] = i
				}
			} else if char == "/" {
				nextStop += 1
			} else {
				nextStop += 9 - strings.Index(RANK, char)
			}
		} 
	}

	board.sidetomove = int8(BOOL(fenparts[1] == "w"))

	board.castleRights[0][0] = strings.Index(fenparts[2], "q") > -1
	board.castleRights[0][1] = strings.Index(fenparts[2], "k") > -1
	board.castleRights[1][0] = strings.Index(fenparts[2], "Q") > -1
	board.castleRights[1][1] = strings.Index(fenparts[2], "K") > -1

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

func (board *Board) PawnDefends(sq int, attacker int8) int {
	return BOOL(board.squares[sq + W - ADVANCES[attacker]] == 2 + attacker) + BOOL(board.squares[sq + E - ADVANCES[attacker]] == 2 + attacker)
}

func (board *Board) Generate(capturesOnly bool) ([]Move, int) {
	attention := &weight_table[board.sidetomove][(board.kings[1-board.sidetomove]%10-1)/2]

	moves := []Move{}

	advance := ADVANCES[board.sidetomove]
	dynamicMobility := 0
	pawnIndexes := make([]int8, 0, 16)

	for i:=A8;i<=H1;i++ {
		piece := board.squares[i]
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
				victim := board.squares[pawnCapture]
				if victim > 1 && (victim&1 != piece&1) {
					moves = append(moves, Move{int8(i), int8(pawnCapture)})
					dynamicMobility += e_pawnattacked * e_divider * BOOL(victim > 3)
				}
			}
		} else {
			ray := (piecetype / 3) == 1
			pattern := patterns[piecetype]
			
			for _, dir := range pattern {
				for end := i + dir;; end += dir {
					victim := board.squares[end]

					if victim == 0 || (victim > 1 && victim&1 != piece&1) {
						if board.PawnDefends(end, 1-board.sidetomove) == 0 {
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
		dynamicMobility += (e_mobility[piecetype] * mobValue)
	}

	if board.enpassant != 0 {
		for _, enpassantStart := range []int{board.enpassant - advance + W, board.enpassant - advance + E} {
			if board.squares[enpassantStart] == 2 + board.sidetomove {
				moves = append(moves, Move{int8(enpassantStart), int8(board.enpassant)})
			}
		}
	}

	// TODO: we could save a lot of time if we moved the incheck inside of the capturesOnly clause since we dont yet use it in eval, and its worthless in qsearch
	kingIndex := board.kings[board.sidetomove]
	board.inCheck = board.attacked(kingIndex, 1-board.sidetomove)
	if !capturesOnly && (kingIndex == E8 || kingIndex == E1) && !board.inCheck {
		if board.castleRights[board.sidetomove][1] && board.squares[kingIndex+E+E] == 0 && board.squares[kingIndex+E] == 0 {
			moves = append(moves, Move{int8(kingIndex), int8(kingIndex+E+E)})
		} 
		if board.castleRights[board.sidetomove][0] && board.squares[kingIndex+W+W+W] == 0 && board.squares[kingIndex+W+W] == 0 && board.squares[kingIndex+W] == 0 {
			moves = append(moves, Move{int8(kingIndex), int8(kingIndex+W+W)})
		}
	}

	board.mobilities[board.sidetomove] = dynamicMobility / e_divider
	score, dynamic_score, dynamics_weight := 0,board.mobilities[1] - board.mobilities[0],0
	dynamic_score += (e_contempt * BOOL((board.ply - int(board.sidetomove)) % 2 == 1)) - (e_contempt * BOOL((board.ply - int(board.sidetomove)) % 2 == 0))  


	for piecetype := range 7 {
		score += e_material[piecetype] * int((board.pieceCount[piecetype * 2 + 1] - board.pieceCount[piecetype * 2]))
		dynamic_score += e_altmaterial[piecetype] * int((board.pieceCount[piecetype * 2 + 1] - board.pieceCount[piecetype * 2]))
		dynamics_weight += phaseWeights[piecetype]* int((board.pieceCount[piecetype * 2 + 1] + board.pieceCount[piecetype * 2]))
	}

	whiterear := [10]int{0,0,0,0,0,0,0,0,0,0,}
	blackrear := [10]int{7,7,7,7,7,7,7,7,7,7,}

	for _, sq := range pawnIndexes {
		piece := board.squares[sq]
		pawnfile := int(sq) % 10
		pawnrank := (int(sq) / 10) - 2

		if (piece & 1) == 1 {
			whiterear[pawnfile] = max(whiterear[pawnfile], pawnrank)
		} else {
			blackrear[pawnfile] = min(blackrear[pawnfile], pawnrank)
		}
	}

	score += e_egKingFile[board.kings[1]%10] + e_egKingRank[board.kings[1]/10-2]
	score -= e_egKingFile[board.kings[0]%10] + e_egKingRank[7-(board.kings[0]/10-2)]
	dynamic_score += e_mgKingFile[board.kings[1]%10] + e_mgKingRank[board.kings[1]/10-2]
	dynamic_score -= e_mgKingFile[board.kings[0]%10] + e_mgKingRank[7-(board.kings[0]/10-2)]

	wkingfile := board.kings[1] % 10
	bkingfile := board.kings[0] % 10

	for i := -1; i <= 1; i++ {
		dynamic_score += e_shieldbase[wkingfile + i] * BOOL(whiterear[wkingfile + i] == 6 || board.squares[board.kings[1] + N + i] == 7)
		dynamic_score += e_shield[wkingfile + i]     * BOOL(whiterear[wkingfile + i] != 0)
		dynamic_score -= e_shieldbase[bkingfile + i] * BOOL(blackrear[bkingfile + i] == 1 || board.squares[board.kings[0] + S + i] == 6)
		dynamic_score -= e_shield[bkingfile + i]     * BOOL(blackrear[bkingfile + i] != 7)
	}
	
	var whitepasser [10]int
	var blackpasser [10]int

	// Weak pawn evaluation
	for _, sq := range pawnIndexes {
		piece := board.squares[sq]
		pfile := int(sq) % 10
		prank := int(sq) / 10 - 2

		if piece & 1 == 1 {
			semiopen := BOOL(blackrear[pfile] == 7) // TODO: can probably inline semi open
			if blackrear[pfile - 1] >= prank && blackrear[pfile] >= prank && blackrear[pfile + 1] >= prank {
				whitepasser[pfile] = max(whitepasser[pfile], 7-prank)
			}

			score += e_phalanx[semiopen] * (BOOL(piece == board.squares[sq+W]) + BOOL(piece == board.squares[sq+E]) + 2 * board.PawnDefends(int(sq), 1))
		} else {
			semiopen := BOOL(whiterear[pfile] == 0)
			if whiterear[pfile - 1] <= prank && whiterear[pfile] <= prank && whiterear[pfile + 1] <= prank {
				blackpasser[pfile] = max(blackpasser[pfile], prank)
			}

			score -= e_phalanx[semiopen] * (BOOL(piece == board.squares[sq+W]) + BOOL(piece == board.squares[sq+E]) + 2 * board.PawnDefends(int(sq), 0))
		}
	}

	// Passer evaluation
	for file := range 10 {
		if whitepasser[file] != 0 { // TODO: we can abuse BOOL here
			score += e_passerRank[whitepasser[file]]
		}
		if blackpasser[file] != 0 {
			score -= e_passerRank[blackpasser[file]]
		}
	}

	final_score := score + (dynamic_score * max(0,dynamics_weight)) / e_divider

	leadingpawns := board.pieceCount[2 + BOOL(final_score >= 0)] // TODO: inline this

	drawishness_weight := e_risk[leadingpawns] // TODO: we can just inline this line lol
	final_score = (final_score * e_divider) / (e_divider + dynamics_weight + drawishness_weight) 
	if board.sidetomove == 0 {
		final_score = -final_score
	}
	return moves, final_score + e_tempo // TODO: maybe we should properly address tempo in tuning
}

func (board *Board) attacked(start int, attacker int8) bool {
	if board.PawnDefends(start, attacker) > 0 {
		return true
	}

	for i, dir := range []int{N,S,E,W,N+W,N+E,S+E,S+W} {
		for sq := start + dir;; sq += dir {
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
			copyBoard.castleRights[copyBoard.sidetomove][0] = false
			copyBoard.castleRights[copyBoard.sidetomove][1] = false
		}

		copyBoard.castleRights[0][0] = copyBoard.castleRights[0][0] && !(move.end == A8 || move.start == A8)
		copyBoard.castleRights[0][1] = copyBoard.castleRights[0][1] && !(move.end == H8 || move.start == H8)
		copyBoard.castleRights[1][0] = copyBoard.castleRights[1][0] && !(move.end == A1 || move.start == A1)
		copyBoard.castleRights[1][1] = copyBoard.castleRights[1][1] && !(move.end == H1 || move.start == H1)
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
		if i % 10 == 9 {
			fmt.Println("")
		}
	}
	fmt.Println(board.zobrist)
}

func (board *Board) Hash() uint64 {
	return board.zobrist ^ Zobrist[15][board.enpassant] ^ Zobrist[15][board.sidetomove] ^ Zobrist[15][2+BOOL(board.castleRights[0][0])] ^ Zobrist[15][4+BOOL(board.castleRights[0][1])] ^ Zobrist[15][6+BOOL(board.castleRights[1][0])] ^ Zobrist[15][8+BOOL(board.castleRights[1][1])]
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
			if -alphabeta(board.Apply(Move{}), -beta, -alpha, ((depth - 4) - (depth / 5)) - ((staticEval - beta)/200), false) >= beta {
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

	if legals == 0 && depth > 0 && !board.inCheck {
		return 0
	}

	bestScore += BOOL(bestScore < -9000) // Mate distance/delay adjustment

	if bestMove.start != bestMove.end { // Many nodes in Qsearch either have no captures, or (in check) no legal captures, not worth storing them in tt since they are cheap and plentiful
		table[hash % hashsize] = entry{hash, bestMove, int16(bestScore), int8(depth), boundtype}
	}

	return bestScore
}

var uciBoard Board
var repetition []uint64 = []uint64{}
var openingBook = map[uint64][]string{}
func printpv(maxdepth int) string {
	pvstr := ""
	board := &uciBoard
	for range maxdepth + 4 {
		if board == nil {
			break
		}
		m := table[board.Hash() % hashsize].move
		if board.Hash() != table[board.Hash() % hashsize].key || m.end == m.start {
			break
		}

		pvstr += m.stringify(board) + " "
		board = board.Apply(m)
	}
	return strings.TrimSpace(pvstr)
}

func main() {
	fmt.Println("info string Started")
	for i := range 15 {
		for j := range 120 {
			Zobrist[i+1][j] = rand.Uint64()
		}
	}

	for a := range e_table {
		mailbox_index := ((a>>3 + 2) * 10) + (a & 7) + 1
		for b := range 4 {
			weight_table[1][b][mailbox_index] = (b * e_table[a]     + (3-b) * e_table[a^7])     / 3
			weight_table[0][b][mailbox_index] = (b * e_table[a^56] + (3-b) * e_table[a^56^7]) / 3
		}
	}

	file, err := os.Open("book2.txt")
    if err == nil {
    	scanner := bufio.NewScanner(file)
        for scanner.Scan() {
        	text := string(scanner.Text()) 
        	if len(text) > 2 && (string(text[0:2]) == "w " || string(text[0:2]) == "b ") {
				bookBoard := FromFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
		    	for i, movestr := range strings.Fields(text)[1:] {
		    		if i % 2 == BOOL(string(text[0]) == "b") {
		    			openingBook[bookBoard.Hash()] = append(openingBook[bookBoard.Hash()], movestr) // Isnt it beautiful that golangs datastructure are useful even with zero values *chefs kiss*
		    		}
		    		bookBoard = *bookBoard.Apply(Move{int8(Parse(movestr[0:2])), int8(Parse(movestr[2:4]))})
		    	}
        	}
        }
    }
    file.Close()

	uciBoard = FromFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	reader := bufio.NewReader(os.Stdin)

	for {
		line, _ := reader.ReadString('\n')
		args := strings.Fields(strings.Replace(line, "startpos", "fen rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 1))

		if len(args) != 0 {
			switch string(args[0]) {
			case "uci":
				fmt.Println("id name Tabby\nid author ffloof\nuciok")
			case "isready":
				fmt.Println("readyok")
			case "ucinewgame":
				history = [2][14][128]int{}
			case "print":
				uciBoard.print()
				uciBoard.sidetomove = 1 - uciBoard.sidetomove
				uciBoard.Generate(true)
				uciBoard.sidetomove = 1 - uciBoard.sidetomove
				_, e := uciBoard.Generate(true)
				fmt.Println("eval", e, "\n")
			case "position":
				uciBoard = FromFen(strings.Join(findAfter("fen", args)[0:4], " "))
				for _, movestr := range findAfter("moves", args) {
					uciBoard = *(uciBoard.Apply(Move{int8(Parse(movestr[0:2])), int8(Parse(movestr[2:4]))}))
				}
				uciBoard.ply = 0
			case "go":
				bookmoves, inbook := openingBook[uciBoard.Hash()]
				if inbook{
					fmt.Println("bestmove", bookmoves[rand.Intn(len(bookmoves))])
					continue
				}

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
				chosenMove := table[uciBoard.Hash() % hashsize].move
				
				streak := 0.0
				for depth := 1; depth <= 100; depth++ {
					fmt.Println("info score cp", alphabeta(&uciBoard, -10000, 10000, depth, true), "depth", depth, "time", time.Now().UnixMilli() - start, "nodes", nodes, "pv", printpv(depth))

					streak *= 0.666
					if chosenMove.start != table[uciBoard.Hash() % hashsize].move.start || chosenMove.end != table[uciBoard.Hash() % hashsize].move.end {
						streak += 0.4
					} 
					chosenMove = table[uciBoard.Hash() % hashsize].move

					if time.Now().UnixMilli() - start > int64(float64(timeAlloc) * (1+streak)) {
						break
					}
				}

				fmt.Println("bestmove", chosenMove.stringify(&uciBoard))				
			case "quit":
				return
			}
		}
	}
}

// Should make a stream where people vote on best move4
// TODO: simplify eval even more? merge some features?
// A. - for example having one pawn shield term would be nice
// B. - similarly not having 4 terms for king position would be nice, in addition to two passer distance terms
// - ^ srsly like half the eval terms are this
// D. phalanx and chain could be unified
// F. phase can just be done with material
// TODO: spsa tune search params (make sure to get time usage as well)
// Should probably retest all the passer and eg king pos terms since they were bugged prior to mailbox update

// Material        = 2 -> 1 term
// Passed Pawns    = 1 term
// Mobility        = 1 term
// Pawn Structure  = 2 -> 1 term
// King Safety     = 6 -> 2 terms (This is the hardest)
// Drawishness?    = 1 term
// Threats?        = 1 term (we could try to remove this term by treating pawn captures as mobility?, or maybe as part of tempo?)
// Tempo?          = 1 term 