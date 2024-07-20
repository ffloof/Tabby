const N = -16
const S = 16
const W = -1
const E = 1

const A8 = 0
const H8 = 7
const A1 = 112
const H1 = 119

const rays = {
    "n": [N + N + W, N + N + E, S + S + W, S + S + E, W + W + N, W + W + S, E + E + N, E + E + S],
    "N": [N + N + W, N + N + E, S + S + W, S + S + E, W + W + N, W + W + S, E + E + N, E + E + S],
    "b": [N + W, N + E, S + W, S + E],
    "B": [N + W, N + E, S + W, S + E],
    "r": [N, S, E, W],
    "R": [N, S, E, W],
    "q": [N, S, E, W, N + W, N + E, S + W, S + E],
    "Q": [N, S, E, W, N + W, N + E, S + W, S + E],
    "k": [N, S, E, W, N + W, N + E, S + W, S + E],
    "K": [N, S, E, W, N + W, N + E, S + W, S + E],
}
an
const sliders = {
    "n": false,
    "N": false,
    "b": true,
    "B": true,
    "r": true,
    "R": true,
    "q": true,
    "Q": true,
    "k": false,
    "K": false,
}

pieceValues = {
    "p": -100,
    "P": 100,
    "n": -300,
    "N": 300,
    "b": -300,
    "B": 300,
    "r": -500,
    "R": 500,
    "q": -900,
    "Q": 900,
    "k": 0,
    "K": 0,
    " ": 0,
    "\n": 0,
    undefined: 0,
}

psqtPawn = [

]
psqtKnight = []
psqtBishop = []
psqtRook = []
psqtQueen = []
psqtKing = []



function sqstr(n) {
    let rank = n >> 4
    let file = n & 7
    return "abcdefgh"[file] + "87654321"[rank]
}

lookup = [
    N + W, 0, 0, 0, 0, 0, 0, N, 0, 0, 0, 0, 0, 0, N + E, 0,
    0, N + W, 0, 0, 0, 0, 0, N, 0, 0, 0, 0, 0, N + E, 0, 0,
    0, 0, N + W, 0, 0, 0, 0, N, 0, 0, 0, 0, N + E, 0, 0, 0,
    0, 0, 0, N + W, 0, 0, 0, N, 0, 0, 0, N + E, 0, 0, 0, 0,
    0, 0, 0, 0, N + W, 0, 0, N, 0, 0, N + E, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, N + W, 0, N, 0, N + E, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, N + W, N, N + E, 0, 0, 0, 0, 0, 0, 0,
    W, W, W, W, W, W, W, 0, E, E, E, E, E, E, E, 0,
    0, 0, 0, 0, 0, 0, S + W, S, S + E, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, S + W, 0, S, 0, S + E, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, S + W, 0, 0, S, 0, 0, S + E, 0, 0, 0, 0, 0,
    0, 0, 0, S + W, 0, 0, 0, S, 0, 0, 0, S + E, 0, 0, 0, 0,
    0, 0, S + W, 0, 0, 0, 0, S, 0, 0, 0, 0, S + E, 0, 0, 0,
    0, S + W, 0, 0, 0, 0, 0, S, 0, 0, 0, 0, 0, S + E, 0, 0,
    S + W, 0, 0, 0, 0, 0, 0, S, 0, 0, 0, 0, 0, 0, S + E, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
]

function initZobrist(n) {
    let byteArray = new BigUint64Array(n)
    return crypto.getRandomValues(byteArray)
}

const zobristTable = {
    "p": initZobrist(128),
    "P": initZobrist(128),
    "n": initZobrist(128),
    "N": initZobrist(128),
    "b": initZobrist(128),
    "B": initZobrist(128),
    "r": initZobrist(128),
    "R": initZobrist(128),
    "q": initZobrist(128),
    "Q": initZobrist(128),
    "k": initZobrist(128),
    "K": initZobrist(128),
    " ": new BigUint64Array(128).fill(0n),
    "\n": new BigUint64Array(128).fill(0n),
}

const zobristSTM = initZobrist(2)
const zobristCastling = initZobrist(8)
const zobristEnpassant = initZobrist(128)

class Move {
    constructor(start, end) {
        this.start = start
        this.end = end
    }

    str() {
        return sqstr(this.start) + sqstr(this.end)
    }

    equals(otherMove) {
        if (otherMove == null) return false
        return otherMove.start == this.start && otherMove.end == this.end
    }
}

class Board {
    constructor(fen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1") {
        if (fen.length == 0) return

        this.incremental = 0
        this.zobrist = 0n
        this.squares = new Array(128 + 1).join(" ")

        let parts = fen.split(" ")

        let i = 0
        for (let char of parts[0].replaceAll("/", "       \n").replaceAll("1", " ").replaceAll("2", "  ").replaceAll("3", "   ").replaceAll("4", "    ").replaceAll("5", "     ").replaceAll("6", "      ").replaceAll("7", "       ").replaceAll("8", "        ")) {
            this.edit(i, char)
            i++
        }

        this.whiteToMove = parts[1] == "w"
        this.enpassant = 0
        this.kings = [this.squares.indexOf("k"), this.squares.indexOf("K")]
        this.castlingRights = [[parts[2].indexOf('q') > -1, parts[2].indexOf('k') > -1], [parts[2].indexOf('Q') > -1, parts[2].indexOf('K') > -1]]
    }

    isHomeRow(i) {
        if (this.whiteToMove) return (A1 + N <= i) && (i <= H1 + N);
        return (A8 + S <= i) && (i <= H8 + S)
    }

    pieceMoves(moves, start, directions, slider) {
        for (let dir of directions) {
            for (let end = start + dir; (end & 0x88) == 0; end += dir) {
                let victim = this.squares[end]

                if (victim != " ") {
                    if ((victim < "a") != this.whiteToMove) moves.push(new Move(start, end))
                    break
                }

                moves.push(new Move(start, end))
                if (!slider) break
            }
        }
    }

    scan(start, directions, slider, checker) {

        for (let dir of directions) {
            for (let end = start + dir; (end & 0x88) == 0; end += dir) {
                let victim = this.squares[end]

                if (victim != " ") {
                    if (victim == checker) return true
                    break
                }


                if (!slider) break
            }
        }
        return false
    }

    // TODO: implement faster cached version
    attacked(index) {
        let pieces = "NBRQK"
        if (this.whiteToMove) {
            pieces = "nbrqk"
            if (this.scan(index, [N + W, N + E], false, "p")) return true
        } else {
            if (this.scan(index, [S + W, S + E], false, "P")) return true
        }

        for (let checkingPiece of pieces) {
            if (this.scan(index, rays[checkingPiece], sliders[checkingPiece], checkingPiece)) return true
        }

        return false
    }

    getHash() {
        return this.zobrist ^
            zobristSTM[Number(this.whiteToMove)] ^
            zobristEnpassant[Number(this.enpassant)] ^
            zobristCastling[0 + Number(this.castlingRights[0][0])] ^
            zobristCastling[2 + Number(this.castlingRights[0][1])] ^
            zobristCastling[4 + Number(this.castlingRights[1][0])] ^
            zobristCastling[6 + Number(this.castlingRights[1][1])]
    }

    generateLegalMoves() {
        let moves = []
        let advance = S
        let pawntype = "p"
        if (this.whiteToMove) {
            pawntype = "P"
            advance = N
        }

        // Standard Moves
        for (let i = 0; i < 120; i++) {
            if ((i & 0x88) != 0) continue

            let piece = this.squares[i]
            if (piece == " ") continue
            if ((piece < "a") != this.whiteToMove) continue

            if (piece == pawntype) {
                if (this.squares[i + advance] == " ") {
                    moves.push(new Move(i, i + advance))
                    if (this.isHomeRow(i) && this.squares[i + advance + advance] == " ") moves.push(new Move(i, i + advance + advance))
                }

                for (let pawnCapture of [i + advance + W, i + advance + E]) {
                    let victim = this.squares[pawnCapture]
                    if (victim <= " ") continue
                    if ((victim < "a") != this.whiteToMove) moves.push(new Move(i, pawnCapture))
                }
            } else {
                this.pieceMoves(moves, i, rays[piece], sliders[piece])
            }
        }


        // Enpassant
        if (this.enpassant > 0) {
            if (this.squares[this.enpassant - advance + W] == pawntype) moves.push(new Move(this.enpassant - advance + W, this.enpassant))
            if (this.squares[this.enpassant - advance + E] == pawntype) moves.push(new Move(this.enpassant - advance + E, this.enpassant))
        }


        let kingIndex = this.kings[Number(this.whiteToMove)]
        this.inCheck = this.attacked(kingIndex)

        // Castling 

        if (!this.inCheck) {
            let rights = this.castlingRights[Number(this.whiteToMove)]
            if (rights[1] && this.squares[kingIndex + E] == " " && this.squares[kingIndex + E + E] == " ") moves.push(new Move(kingIndex, kingIndex + E + E))
            if (rights[0] && this.squares[kingIndex + W] == " " && this.squares[kingIndex + W + W] == " " && this.squares[kingIndex + W + W + W] == " ") moves.push(new Move(kingIndex, kingIndex + W + W))
        }

        return moves
    }

    edit(i, newpiece) {
        // Arrays are immutable in js sadge
        let oldpiece = this.squares[i]

        this.zobrist ^= zobristTable[oldpiece][i]
        this.zobrist ^= zobristTable[newpiece][i]
        this.squares = this.squares.substring(0, i) + newpiece + this.squares.substring(i + 1);


        this.incremental += pieceValues[newpiece] - pieceValues[oldpiece]
    }

    apply(move) {
        let copyBoard = new Board("")
        // Object.assign(copyBoard, this) <- would work we have to be careful with removing check tho

        copyBoard.squares = this.squares
        copyBoard.incremental = this.incremental
        copyBoard.zobrist = this.zobrist

        let movingPiece = this.squares[move.start]

        copyBoard.edit(move.start, " ")
        copyBoard.edit(move.end, movingPiece)


        copyBoard.whiteToMove = this.whiteToMove
        copyBoard.enpassant = 0
        copyBoard.kings = this.kings.slice()
        copyBoard.castlingRights = [this.castlingRights[0].slice(), this.castlingRights[1].slice()]


        if (movingPiece == "p") {
            if (move.end >= A1) copyBoard.edit(move.end, "q")
            if (move.end - move.start == S + S) copyBoard.enpassant = move.start + S
            if (move.end == this.enpassant) copyBoard.edit(move.end + N, " ")
        } else if (movingPiece == "P") {
            if (move.end <= H8) copyBoard.edit(move.end, "Q")
            if (move.end - move.start == N + N) copyBoard.enpassant = move.start + N
            if (move.end == this.enpassant) copyBoard.edit(move.end + S, " ")
        } else if (movingPiece == "k" || movingPiece == "K") {
            copyBoard.kings[Number(this.whiteToMove)] = move.end
            copyBoard.castlingRights[Number(this.whiteToMove)] = [false, false]
            let coloredRook = "r"
            if (this.whiteToMove) coloredRook = "R"

            if (move.end - move.start == W + W) {
                if (copyBoard.attacked(move.start + W)) return null
                copyBoard.edit(move.end + E, coloredRook)
                copyBoard.edit(move.end + W + W, " ")
            }
            if (move.end - move.start == E + E) {
                if (copyBoard.attacked(move.start + E)) return null
                copyBoard.edit(move.end + W, coloredRook)
                copyBoard.edit(move.end + E, " ")
            }
        }

        if (copyBoard.attacked(copyBoard.kings[Number(this.whiteToMove)])) return null

        if (move.start == H1 || move.end == H1) copyBoard.castlingRights[1][1] = false
        if (move.start == H8 || move.end == H8) copyBoard.castlingRights[0][1] = false
        if (move.start == A1 || move.end == A1) copyBoard.castlingRights[1][0] = false
        if (move.start == A8 || move.end == A8) copyBoard.castlingRights[0][0] = false

        copyBoard.whiteToMove = !copyBoard.whiteToMove
        return copyBoard
    }
}

function perft(perftboard, depth) {
    if (depth == 0) return 1

    let movelist = perftboard.generateLegalMoves()
    let nodes = 0

    for (let move of movelist) {
        let nextBoard = perftboard.apply(move)
        if (nextBoard != null) nodes += perft(nextBoard, depth - 1)
    }


    return nodes
}

function tester(depth, fenstr = "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8") {
    start = Date.now()

    let totalNodes = 0
    let movecount = 0

    testboard = new Board(fenstr)
    let testmoves = testboard.generateLegalMoves()
    for (let move of testmoves) {
        let perftboard = testboard.apply(move)
        if (perftboard != null) {
            movecount += 1
            let nodes = perft(perftboard, depth - 1)
            totalNodes += nodes
            try {
                console.log(move.str(), nodes)
            } catch (err) {
                console.log(move)

            }
        }
    }

    console.log(totalNodes, "nodes", movecount, "moves", Date.now() - start, "ms")
}

tester(4)



const MVVLVA = { "p": 100, "P": 100, "n": 300, "N": 300, "b": 300, "B": 300, "r": 500, "R": 500, "q": 900, "Q": 900, "k": 999, "K": 999, " ": -100 }

class TableEntry {
    constructor(hash, bestmove) {
        this.hash = hash
        this.bestmove = bestmove
    }
}

class TranspositionTable {
    constructor() {
        this.size = BigInt(1 << 20)
        this.table = new Array(this.size)
    }

    get(hash) {
        let entry = this.table[hash % this.size]
        if (entry == null) return null
        if (entry.hash == hash) return entry
        return null
    }

    set(hash, bestmove) {
        this.table[hash % this.size] = new TableEntry(hash, bestmove)
    }
}

function eval(board) {
    let score = board.incremental
    if (board.whiteToMove) return score
    return -score
}


const globalTT = new TranspositionTable()
var nodes = 0

function alphabeta(board, depth, alpha = -10000, beta = 10000) {
    let hash = board.getHash()

    let bestScore = -9999
    if (depth <= 0) {
        depth = 0
        bestScore = eval(board);
        if (bestScore > alpha) alpha = bestScore
        if (alpha >= beta) return bestScore

    }

    let tableMove = null
    let entry = globalTT.get(hash)
    if (entry) {
        tableMove = entry.tableMove
    }


    let bestmove = null
    let moves = board.generateLegalMoves()

    let heuristic = new Array(moves.length).fill(0)
    for (let i in moves) {
        heuristic[i] = 1000 * MVVLVA[board.squares[moves[i].end]] - MVVLVA[board.squares[moves[i].start]]
        if (moves[i].equals(tableMove)) heuristic[i] = 10000000
    }

    let indices = [...Array(moves.length).keys()];
    indices.sort((a, b) => heuristic[b] - heuristic[a]);
    moves = indices.map(i => moves[i]);

    for (let move of moves) {
        if (depth <= 0 && board.squares[move.end] == " ") continue
        let nextboard = board.apply(move)

        if (nextboard == null) continue
        nodes++
        let score = -alphabeta(nextboard, depth - 1, -beta, -alpha)

        if (score > bestScore) {
            bestScore = score
            bestmove = move
        }
        if (score > alpha) alpha = score
        if (score >= beta) break

    }

    globalTT.set(hash, bestmove)

    return bestScore
}



//tester(4)

let board = new Board("r1bqkb1r/ppp2ppp/2n5/3np1N1/2B5/8/PPPP1PPP/RNBQK2R w KQkq - 0 6")

console.log()

function iterativeDeepening(board) {
    nodes = 0
    for (let depth = 2; depth < 128; depth++) {
        let start = new Date()
        alphabeta(board, depth)
        let end = new Date()
        console.log(globalTT.get(board.getHash()).bestmove.str(), nodes, end - start)
        if (end - start > 1000) break
    }
}

iterativeDeepening(board)