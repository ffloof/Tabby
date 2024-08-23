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

psqt = {
    "P": [
        106,106,106,106,106,106,106,106,  0,0,0,0,0,0,0,0,
        283,263,256,235,240,238,273,263,  0,0,0,0,0,0,0,0,
        177,186,175,154,164,166,181,174,  0,0,0,0,0,0,0,0,
        125,121,107,111,113,109,120,111,  0,0,0,0,0,0,0,0,
        95,107,96,112,105,95,108,88,  0,0,0,0,0,0,0,0,
        95,106,98,96,106,92,118,96,  0,0,0,0,0,0,0,0,
        94,111,91,80,86,111,122,94,  0,0,0,0,0,0,0,0,
        106,106,106,106,106,106,106,106,  0,0,0,0,0,0,0,0,
    ],
    "p":[
        -106,-106,-106,-106,-106,-106,-106,-106,  0,0,0,0,0,0,0,0,
        -94,-111,-91,-80,-86,-111,-122,-94,  0,0,0,0,0,0,0,0,
        -95,-106,-98,-96,-106,-92,-118,-96,  0,0,0,0,0,0,0,0,
        -95,-107,-96,-112,-105,-95,-108,-88,  0,0,0,0,0,0,0,0,
        -125,-121,-107,-111,-113,-109,-120,-111,  0,0,0,0,0,0,0,0,
        -177,-186,-175,-154,-164,-166,-181,-174,  0,0,0,0,0,0,0,0,
        -283,-263,-256,-235,-240,-238,-273,-263,  0,0,0,0,0,0,0,0,
        -106,-106,-106,-106,-106,-106,-106,-106,  0,0,0,0,0,0,0,0,
    ],
    "N": [
        217,248,336,318,345,284,269,253,  0,0,0,0,0,0,0,0,
        317,335,397,355,385,387,344,319,  0,0,0,0,0,0,0,0,
        342,361,391,410,410,413,380,373,  0,0,0,0,0,0,0,0,
        353,376,381,407,386,392,357,364,  0,0,0,0,0,0,0,0,
        347,360,378,375,388,378,366,337,  0,0,0,0,0,0,0,0,
        341,352,367,374,383,364,368,331,  0,0,0,0,0,0,0,0,
        318,338,337,353,354,361,342,336,  0,0,0,0,0,0,0,0,
        295,325,326,334,326,332,319,328,  0,0,0,0,0,0,0,0,
    ],
    "n": [
        -295,-325,-326,-334,-326,-332,-319,-328,  0,0,0,0,0,0,0,0,
        -318,-338,-337,-353,-354,-361,-342,-336,  0,0,0,0,0,0,0,0,
        -341,-352,-367,-374,-383,-364,-368,-331,  0,0,0,0,0,0,0,0,
        -347,-360,-378,-375,-388,-378,-366,-337,  0,0,0,0,0,0,0,0,
        -353,-376,-381,-407,-386,-392,-357,-364,  0,0,0,0,0,0,0,0,
        -342,-361,-391,-410,-410,-413,-380,-373,  0,0,0,0,0,0,0,0,
        -317,-335,-397,-355,-385,-387,-344,-319,  0,0,0,0,0,0,0,0,
        -217,-248,-336,-318,-345,-284,-269,-253,  0,0,0,0,0,0,0,0,
    ],
    "B": [
        355,343,372,369,373,363,354,363,  0,0,0,0,0,0,0,0,
        360,401,382,383,405,403,398,385,  0,0,0,0,0,0,0,0,
        385,398,416,410,397,415,409,408,  0,0,0,0,0,0,0,0,
        379,385,394,409,414,391,393,382,  0,0,0,0,0,0,0,0,
        372,395,404,406,416,397,379,360,  0,0,0,0,0,0,0,0,
        387,398,401,396,399,402,398,383,  0,0,0,0,0,0,0,0,
        380,394,382,385,386,388,404,381,  0,0,0,0,0,0,0,0,
        375,372,372,370,366,353,381,359,  0,0,0,0,0,0,0,0,
    ],
    "b": [
        -375,-372,-372,-370,-366,-353,-381,-359,  0,0,0,0,0,0,0,0,
        -380,-394,-382,-385,-386,-388,-404,-381,  0,0,0,0,0,0,0,0,
        -387,-398,-401,-396,-399,-402,-398,-383,  0,0,0,0,0,0,0,0,
        -372,-395,-404,-406,-416,-397,-379,-360,  0,0,0,0,0,0,0,0,
        -379,-385,-394,-409,-414,-391,-393,-382,  0,0,0,0,0,0,0,0,
        -385,-398,-416,-410,-397,-415,-409,-408,  0,0,0,0,0,0,0,0,
        -360,-401,-382,-383,-405,-403,-398,-385,  0,0,0,0,0,0,0,0,
        -355,-343,-372,-369,-373,-363,-354,-363,  0,0,0,0,0,0,0,0,
    ],
    "R": [
        603,598,605,597,602,587,591,582,  0,0,0,0,0,0,0,0,
        601,602,607,611,601,602,597,586,  0,0,0,0,0,0,0,0,
        592,588,592,595,599,596,586,601,  0,0,0,0,0,0,0,0,
        581,583,589,598,585,590,582,586,  0,0,0,0,0,0,0,0,
        578,572,574,586,581,585,569,560,  0,0,0,0,0,0,0,0,
        566,568,565,575,566,566,557,554,  0,0,0,0,0,0,0,0,
        559,578,575,575,564,567,587,537,  0,0,0,0,0,0,0,0,
        559,566,578,579,578,562,545,535,  0,0,0,0,0,0,0,0,
    ],
    "r": [
        -559,-566,-578,-579,-578,-562,-545,-535,  0,0,0,0,0,0,0,0,
        -559,-578,-575,-575,-564,-567,-587,-537,  0,0,0,0,0,0,0,0,
        -566,-568,-565,-575,-566,-566,-557,-554,  0,0,0,0,0,0,0,0,
        -578,-572,-574,-586,-581,-585,-569,-560,  0,0,0,0,0,0,0,0,
        -581,-583,-589,-598,-585,-590,-582,-586,  0,0,0,0,0,0,0,0,
        -592,-588,-592,-595,-599,-596,-586,-601,  0,0,0,0,0,0,0,0,
        -601,-602,-607,-611,-601,-602,-597,-586,  0,0,0,0,0,0,0,0,
        -603,-598,-605,-597,-602,-587,-591,-582,  0,0,0,0,0,0,0,0,
    ],
    "Q": [
        1116,1123,1163,1171,1131,1163,1118,1131,  0,0,0,0,0,0,0,0,
        1107,1105,1130,1135,1129,1209,1162,1177,  0,0,0,0,0,0,0,0,
        1099,1116,1134,1164,1196,1170,1206,1183,  0,0,0,0,0,0,0,0,
        1095,1111,1142,1139,1160,1151,1139,1151,  0,0,0,0,0,0,0,0,
        1107,1125,1131,1131,1143,1135,1149,1110,  0,0,0,0,0,0,0,0,
        1104,1121,1123,1127,1126,1128,1133,1116,  0,0,0,0,0,0,0,0,
        1114,1120,1126,1122,1129,1137,1113,1118,  0,0,0,0,0,0,0,0,
        1134,1106,1114,1120,1103,1088,1065,1094,  0,0,0,0,0,0,0,0,
    ],
    "q": [
        -1134,-1106,-1114,-1120,-1103,-1088,-1065,-1094,  0,0,0,0,0,0,0,0,
        -1114,-1120,-1126,-1122,-1129,-1137,-1113,-1118,  0,0,0,0,0,0,0,0,
        -1104,-1121,-1123,-1127,-1126,-1128,-1133,-1116,  0,0,0,0,0,0,0,0,
        -1107,-1125,-1131,-1131,-1143,-1135,-1149,-1110,  0,0,0,0,0,0,0,0,
        -1095,-1111,-1142,-1139,-1160,-1151,-1139,-1151,  0,0,0,0,0,0,0,0,
        -1099,-1116,-1134,-1164,-1196,-1170,-1206,-1183,  0,0,0,0,0,0,0,0,
        -1107,-1105,-1130,-1135,-1129,-1209,-1162,-1177,  0,0,0,0,0,0,0,0,
        -1116,-1123,-1163,-1171,-1131,-1163,-1118,-1131,  0,0,0,0,0,0,0,0,
    ],
    "K":[
        -53,-5,-36,-4,-6,32,-21,-70,   0,0,0,0,0,0,0,0,
        4,32,-6,5,18,20,12,-22,  0,0,0,0,0,0,0,0,
        -20,6,26,25,24,35,41,5,  0,0,0,0,0,0,0,0,
        -16,0,36,27,31,30,21,-6,  0,0,0,0,0,0,0,0,
        -34,5,26,29,35,23,5,-22,  0,0,0,0,0,0,0,0,
        -22,7,16,18,24,17,4,-26,  0,0,0,0,0,0,0,0,
        -27,-10,0,-7,0,6,2,-26,  0,0,0,0,0,0,0,0,
        -55,-6,-11,-27,-18,-30,-10,-36,   0,0,0,0,0,0,0,0,
    ],
    "k":[
        55,6,11,27,18,30,10,36,   0,0,0,0,0,0,0,0,
        27,10,0,7,0,-6,-2,26,   0,0,0,0,0,0,0,0,
        22,-7,-16,-18,-24,-17,-4,26,   0,0,0,0,0,0,0,0,
        34,-5,-26,-29,-35,-23,-5,22,   0,0,0,0,0,0,0,0,
        16,0,-36,-27,-31,-30,-21,6,   0,0,0,0,0,0,0,0,
        20,-6,-26,-25,-24,-35,-41,-5,   0,0,0,0,0,0,0,0,
        -4,-32,6,-5,-18,-20,-12,22,   0,0,0,0,0,0,0,0,
        53,5,36,4,6,-32,21,70,   0,0,0,0,0,0,0,0,
    ],
    " ": new Int32Array(128).fill(0),
    "\n": new Int32Array(128).fill(0),
    undefined: new Int32Array(128).fill(0),
}

function sqstr(n) {
    let rank = n >> 4
    let file = n & 7
    return "abcdefgh"[file] + "87654321"[rank]
}


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
const zobristCastlingQ = initZobrist(2)
const zobristCastlingK = initZobrist(2)
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
        this.castleQueenside = [parts[2].indexOf('q') > -1, , parts[2].indexOf('Q') > -1]
        this.castleKingside = [parts[2].indexOf('k') > -1, parts[2].indexOf('K') > -1]
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
        let stm = Number(this.whiteToMove)
        let hash = this.zobrist ^ zobristSTM[stm] ^ zobristEnpassant[this.enpassant]
        if (this.castleKingside[0]) hash ^= zobristCastlingK[0]
        if (this.castleQueenside[0]) hash ^= zobristCastlingQ[0]
        if (this.castleKingside[1]) hash ^= zobristCastlingK[1]
        if (this.castleQueenside[1]) hash ^= zobristCastlingQ[1]
        return hash
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
                    if ((pawnCapture & 0x88) != 0) continue
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
            if (this.castleKingside[Number(this.whiteToMove)] && this.squares[kingIndex + E] == " " && this.squares[kingIndex + E + E] == " ") moves.push(new Move(kingIndex, kingIndex + E + E))
            if (this.castleQueenside[Number(this.whiteToMove)] && this.squares[kingIndex + W] == " " && this.squares[kingIndex + W + W] == " " && this.squares[kingIndex + W + W + W] == " ") moves.push(new Move(kingIndex, kingIndex + W + W))
        }

        return moves
    }

    edit(i, newpiece) {
        // Arrays are immutable in js sadge
        let oldpiece = this.squares[i]

        this.zobrist ^= zobristTable[oldpiece][i]
        this.zobrist ^= zobristTable[newpiece][i]
        this.squares = this.squares.substring(0, i) + newpiece + this.squares.substring(i + 1);


        //this.incremental += pieceValues[newpiece] - pieceValues[oldpiece]
        this.incremental += psqt[newpiece][i] - psqt[oldpiece][i]
    }

    apply(move) {
        let copyBoard = new Board("")
        // Object.assign(copyBoard, this) <- would work we have to be careful with removing check tho

        copyBoard.squares = this.squares
        copyBoard.incremental = this.incremental
        copyBoard.zobrist = this.zobrist

        let movingPiece = this.squares[move.start]

        if (move.end == -1) console.log("???", move.start)

        copyBoard.edit(move.start, " ")
        copyBoard.edit(move.end, movingPiece)
        


        copyBoard.whiteToMove = this.whiteToMove
        copyBoard.enpassant = 0
        copyBoard.kings = this.kings.slice()
        copyBoard.castleKingside = this.castleKingside.slice()
        copyBoard.castleQueenside =  this.castleQueenside.slice()


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
            copyBoard.castleKingside[Number(this.whiteToMove)] = false
            copyBoard.castleQueenside[Number(this.whiteToMove)] = false
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

        if (move.start == H1 || move.end == H1) copyBoard.castleKingside[1] = false
        if (move.start == H8 || move.end == H8) copyBoard.castleKingside[0] = false
        if (move.start == A1 || move.end == A1) copyBoard.castleQueenside[1] = false
        if (move.start == A8 || move.end == A8) copyBoard.castleQueenside[0] = false

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

const MVVLVA = { "p": 100, "P": 100, "n": 300, "N": 300, "b": 300, "B": 300, "r": 500, "R": 500, "q": 900, "Q": 900, "k": 999, "K": 999, " ": 0 }

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

// In general I want to have an evaluation function centered around 4 components
// 1. Material
// 2. King Safety
// 3. Activity
// 4. Pawn Structure

// Things I want to try right now
// Pawn/Nonpawn attacks / defence bonuses
// Attack penalty for biting on granite? 

function eval(board) {
    let score = board.incremental
    if (board.whiteToMove) return score
    return -score
}


const globalTT = new TranspositionTable()
var nodes = 0

function swap(list, a, b){
    let temp = list[a]
    list[a] = list[b]
    list[b] = temp
}

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
        heuristic[i] = Math.max(1000 * MVVLVA[board.squares[moves[i].end]] - MVVLVA[board.squares[moves[i].start]], 0)
        if (moves[i].equals(tableMove)) heuristic[i] = 10000000
    }

    for (let i=0;i<moves.length;i++){
        let bestIndex = -1;
        let bestHeuristic = -100000;
        for(let j=i;j<moves.length;j++){
            if (bestHeuristic < heuristic[j]) {
                bestIndex = j
                bestHeuristic = heuristic[j]
            }
        }

        let move = moves[bestIndex]
        swap(moves, i, bestIndex)
        swap(heuristic, i, bestIndex)
        

    //for (let move of moves) {
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

function iterativeDeepening(board) {
    nodes = 0
    for (let depth = 1; depth < 128; depth++) {
        let start = new Date()
        let score = alphabeta(board, depth)
        let end = new Date()
        console.log(depth, score, globalTT.get(board.getHash()).bestmove.str(), nodes, end - start)
        if (end - start > 1000) break
    }
}

function getWordsAfter(base, word) {
    let indexof = base.indexOf(word)
    if (indexof == -1) return ""
    return base.substring(indexof + word.length + 1).split(" ")
}

var uciBoard = new Board()

function parseUCI(ucistr){
    command = ucistr.replace("startpos", "fen rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1").split(" ")[0]
    switch(command){
        case "uci":
            console.log("id name Tabby")
            console.log("id author ffloof")
            console.log("uciok")
            break
        case "isready":
            console.log("readyok")
            break
        case "position":
            uciBoard = new Board(getWordsAfter(ucistr, "position fen").slice(0,4).join(" "))
            let movestrs = getWordsAfter("moves")
            break
        case "go":
            iterativeDeepening(uciBoard)
            break
        case "perft":
            let depth = Number(getWordsAfter(ucistr, "perft")[0])
            start = Date.now()

            let totalNodes = 0
            let movecount = 0

            let testmoves = uciBoard.generateLegalMoves()
            for (let move of testmoves) {
                let perftboard = uciBoard.apply(move)
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
            break
    }
}

parseUCI("position fen r1bq3r/ppp1kppp/2nbpn2/3p4/3P4/2NBPN2/PPP1KPPP/R1BQ3R w - - 6 7")
//parseUCI("go")
parseUCI("perft 4")

function tester(){
    parseUCI("perft 4")
}