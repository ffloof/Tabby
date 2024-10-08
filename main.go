package main

import (
	"fmt"
	"strings"
	"bufio"
	"os"
)

const N, S, E, W = -16, 16, 1, -1
const A8, H8, A1, H1 = 0, 7, 112, 119
const E8, E1 = 4, 116
const CASTLE = E * 8
const PIECE = " .pPnNbBrRqQkK"
const RANK = "87654321"
const FILE = "abcdefgh"
var ADVANCES = [...]int{S,N}

type Board struct {
	squares    [128]int8
	kings      [2]int
	enpassant  int
	sidetomove int8
}

type Move struct {
	start int8
	end   int8
}

func (move Move) stringify() string {
	return string(FILE[move.start&7]) + string(RANK[move.start>>4]) + string(FILE[move.end&7]) + string(RANK[move.end>>4])
}

type Entry struct {
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
	board.squares[index] = newpiece
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


			for _, dir := range pattern {
				for end := i + dir; (end & 0x88) == 0; end += dir {
					victim := board.squares[end]

					if victim != 0 {
						if victim&1 != piece&1 {
							moves = append(moves, Move{int8(i), int8(end)})
						}
					} else if !capturesOnly {
						moves = append(moves, Move{int8(i), int8(end)})
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

var uciBoard Board

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
	case "go":
		for depth := range 100 {
			fmt.Println("info string depth", depth)
		}
	case "perft":
		fmt.Println("total", perft(&uciBoard, 5, 5))
	case "position":
		uciBoard = FromFen(strings.Join(findAfter("fen", args)[0:4], " "))
		for _, movestr := range findAfter("moves", args) {
			uciBoard = *(uciBoard.Apply(Move{int8(Parse(movestr[0:2])), int8(Parse(movestr[2:4]))}))
			// TODO: add underpromotion condition?
		}
	case "go":
	case "quit":
		return
	}
}

func main() {
	uciBoard = FromFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	reader := bufio.NewReader(os.Stdin)

	parseuci("position fen rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8")

	for {
		line, _ := reader.ReadString('\n')
		parseuci(line)
	}
}

// Eval centered around
// 1. Material
//     - static count
// 2. King Safety
//     - pawn shield
//     - king vmob?
// 3. Activity
//     - mobility
//     - attacks?
//     - weighted mobility?
// 4. Pawn Structure / Endgames
//     - drawishness?
//     - backwards pawns?
//     

// Ben finegolds middle name is philip
// Should make a stream where people vote on best move
// TODO: should we make king capture engine? or just regular?