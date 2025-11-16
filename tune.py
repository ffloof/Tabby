import sys
import chess
import numpy as np
import random
from tqdm import tqdm
import matplotlib.pyplot as plt

dataPath = sys.argv[1]

file1 = open(dataPath, 'r')

lines = []

for line in file1.readlines():
    lines.append(line)

random.shuffle(lines)

outcomescore = {
    "\"1-0\";":1,
    "\"1/2-1/2\";":0,
    "\"0-1\";":-1,
    "1-0":1,
    "0-1":-1,
    "1/2-1/2":0,
    "[1.0]":1,
    "[0.0]":-1,
    "[0.5]":0,
}

sign = [-1,1]

inputs = []
outputs = []

N = -10
S = 10
W = -1
E = 1

mailbox = [
    21, 22, 23, 24, 25, 26, 27, 28,
    31, 32, 33, 34, 35, 36, 37, 38,
    41, 42, 43, 44, 45, 46, 47, 48,
    51, 52, 53, 54, 55, 56, 57, 58,
    61, 62, 63, 64, 65, 66, 67, 68,
    71, 72, 73, 74, 75, 76, 77, 78,
    81, 82, 83, 84, 85, 86, 87, 88,
    91, 92, 93, 94, 95, 96, 97, 98
]

inverse = [
    -1,-1,-1,-1,-1,-1,-1,-1,-1,-1,
    -1,-1,-1,-1,-1,-1,-1,-1,-1,-1,
    -1, 0, 1, 2, 3, 4, 5, 6, 7,-1,
    -1, 8, 9,10,11,12,13,14,15,-1,
    -1,16,17,18,19,20,21,22,23,-1,
    -1,24,25,26,27,28,29,30,31,-1,
    -1,32,33,34,35,36,37,38,39,-1,
    -1,40,41,42,43,44,45,46,47,-1,
    -1,48,49,50,51,52,53,54,55,-1,
    -1,56,57,58,59,60,61,62,63,-1,
    -1,-1,-1,-1,-1,-1,-1,-1,-1,-1,
    -1,-1,-1,-1,-1,-1,-1,-1,-1,-1,
]

starts = [0,]
sizes = []
first = True

rays = [ False, False, False, True, True, True, False]
patterns = [ [], [], [N+N+W,N+N+E,S+S+W,S+S+E,W+W+N,W+W+S,E+E+N,E+E+S], [N+W,N+E,S+W,S+E], [N,S,E,W], [N,S,E,W,N+W,N+E,S+W,S+E], [N,S,E,W,N+W,N+E,S+W,S+E]]

lines = lines[:2_000_000]

for line in tqdm(lines):
    if len(outputs) >= len(lines):
        break

    packed = line.split("c9")

    fen = line[:line.rfind(" ")].strip()
    outcome = outcomescore[line[line.rfind(" "):].strip()]

    turn = fen.split(" ")[1].lower().strip() == "w"
    virtualboard = [
        1,1,1,1,1,1,1,1,1,1,
        1,1,1,1,1,1,1,1,1,1,
        1,0,0,0,0,0,0,0,0,1,
        1,0,0,0,0,0,0,0,0,1,
        1,0,0,0,0,0,0,0,0,1,
        1,0,0,0,0,0,0,0,0,1,
        1,0,0,0,0,0,0,0,0,1,
        1,0,0,0,0,0,0,0,0,1,
        1,0,0,0,0,0,0,0,0,1,
        1,0,0,0,0,0,0,0,0,1,
        1,1,1,1,1,1,1,1,1,1,
        1,1,1,1,1,1,1,1,1,1,
    ]

    try:
        i = 0
        c = 0
        while i < 64:
            char = fen[c]
            c += 1
            if char in "pPnNbBrRqQkK":
                virtualboard[mailbox[i]] = "pPnNbBrRqQkK".index(char) + 2
            elif char == "/":
                continue
            elif char in "12345678":
                i += int(char) - 1
            else:
                break
            i += 1
    except Exception as err:
        print("Error:", err)

    material = np.zeros((2, 7), dtype=np.int8)
    sidetomove = np.zeros(1, dtype=np.int8)
    baseMob = np.zeros((2,7), dtype=np.int8)
    phalanx = np.zeros(2, dtype=np.int8)
    chain = np.zeros(2, dtype=np.int8)
    passerDistance = np.zeros(9, dtype=np.int8)
    passerSupportDistance = np.zeros(9, dtype=np.int8)
    
    passerRank = np.array([
        [0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
        [11,11,11,11,11,11,11,11,11,11],
    ], dtype=np.int8)

    shield = np.zeros((2,10), dtype=np.int8)
    shieldbase = np.zeros((2,10), dtype=np.int8)

    mobtable = np.zeros((2,6,64), dtype=np.int8)
    othertable = np.zeros((2,1,64), dtype=np.int8)

    kings = [-1, -1]
    rearpawns = [
        [11,11,11,11,11,11,11,11,11,11],
        [0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    ]
    
    captures = np.zeros((2,7,7), dtype=np.int8)

    for i in range(64):
        piece = virtualboard[mailbox[i]]
        piecetype = piece // 2

        if piecetype == 0:
            continue

        piececolor = piece & 1

        material[piececolor, piecetype] += 1

        if piecetype > 0:
            mobtable[piece&1][piecetype-1][i] += 1

        if piecetype == 1:
            pfile = mailbox[i] % 10
            prank = mailbox[i] // 10

            if piececolor == 0:
                rearpawns[piececolor][pfile] = min(rearpawns[piececolor][pfile], prank)
            else:
                rearpawns[piececolor][pfile] = max(rearpawns[piececolor][pfile], prank)

        elif piecetype == 6:
            kings[piececolor] = mailbox[i]

    wkingfile = kings[1] % 10
    wkingrank = kings[1] // 10
    bkingfile = kings[0] % 10
    bkingrank = kings[0] // 10

    oppcolor = [[False,False],[False,False]]
    
    for sq in range(len(virtualboard)):
        piece = virtualboard[sq]

        if piece < 2:
            continue

        piecetype = piece // 2
        isray = rays[piecetype]
        pattern = patterns[piecetype]

        defended = (piece & 1 == 0 and (virtualboard[sq+N+W] == 2 or virtualboard[sq+N+E] == 2)) or (piece & 1 == 1 and (virtualboard[sq+S+W] == 3 or virtualboard[sq+S+E] == 3))

        if piecetype == 3:
            oppcolor[piece&1][sq&1] = True

        if piecetype == 1:
            pfile = sq % 10
            prank = sq // 10
            
            if piece == 2:
                pattern = [S+W, S+E]

            if piece == 3:
                pattern = [N+W, N+E]


            j = inverse[sq]

            if piece & 1 == 0:
                if virtualboard[sq + W] == 2 or virtualboard[sq + E] == 2:
                    if rearpawns[1][pfile] == 0:
                        phalanx[1] -= 1
                    else:
                        phalanx[0] -= 1

                if virtualboard[sq + W + N] == 2 or virtualboard[sq + E + N] == 2:
                    if rearpawns[1][pfile] == 0:
                        chain[1] -= 1
                    else:
                        chain[0] -= 1

                if rearpawns[1][pfile - 1] <= prank and rearpawns[1][pfile] <= prank and rearpawns[1][pfile + 1] <= prank:
                    passerRank[0][pfile] = max(prank, passerRank[0][pfile])

            else:
                if virtualboard[sq + W] == 3 or virtualboard[sq + E] == 3:
                    if rearpawns[0][pfile] == 11:
                        phalanx[1] += 1
                    else:
                        phalanx[0] += 1

                if virtualboard[sq + W + S] == 3 or virtualboard[sq + E + S] == 3:
                    if rearpawns[0][pfile] == 11:
                        chain[1] += 1
                    else:
                        chain[0] += 1

                if rearpawns[0][pfile - 1] >= prank and rearpawns[0][pfile] >= prank and rearpawns[0][pfile + 1] >= prank:
                    passerRank[1][pfile] = min(prank, passerRank[1][pfile])
        
        
        

        for direction in pattern:
            current = sq
            for i in range(1,8):
                current += direction

                if virtualboard[current] == 1:
                    break

                pawnDefence = False
                
                if piecetype != 1:
                    baseMob[piece&1][piecetype] += 1

                if virtualboard[current] == 0 or ((virtualboard[current] & 1) != (piece & 1)):
                    if piecetype != 1:
                        if piece & 1 == 0 and (virtualboard[current+S+W] == 3 or virtualboard[current+S+E] == 3):
                            pawnDefence = True
                        elif piece & 1 == 1 and (virtualboard[current+N+W] == 2 or virtualboard[current+N+E] == 2):
                            pawnDefence = True
                        
                        if not pawnDefence:
                            mobtable[piece&1][piecetype-1][inverse[current]] += 1

                    if virtualboard[current] != 0 and ((virtualboard[current] & 1) != (piece & 1)):
                        if virtualboard[current] // 2 > 1 or not pawnDefence:
                            captures[piece & 1][piecetype][virtualboard[current] // 2] += 1

                if virtualboard[current] != 0 or (not isray):
                    break

    for x in [-1,0,1]:
        wspawn = rearpawns[1][wkingfile + x]
        if wspawn == 8:
            shieldbase[1][wkingfile + x] = 1
        if wspawn > 0:
            shield[1][wkingfile + x] = 1

        bspawn = rearpawns[0][bkingfile + x]
        if bspawn == 3:
            shieldbase[0][bkingfile + x] = 1
        if bspawn < 10:
            shield[0][bkingfile + x] = 1

    pushers = np.zeros(8, dtype=np.int8)

    BTEMPO = 1
    WTEMPO = 0
    if turn:
        WTEMPO = 1
        BTEMPO = 0


    for i in range(10):
        wPass = 11 - passerRank[1][i]
        bPass = passerRank[0][i]
        wPass -= 2
        bPass -= 2

        if wPass >= 0:
            pushers[wPass] += 1
            passerDistance[max(abs(i-bkingfile),8*int(passerRank[1][i]<bkingrank))] += 1
            passerSupportDistance[max(abs(i-wkingfile),8*int(passerRank[1][i]<wkingrank))] += 1

        if bPass >= 0:
            pushers[bPass] -= 1
            passerDistance[max(abs(i-wkingfile),8*int(passerRank[0][i]>wkingrank))] -= 1
            passerSupportDistance[max(abs(i-bkingfile),8*int(passerRank[0][i]>bkingrank))] -= 1

    sidetomove[0] = sign[turn]

    npawns = np.zeros((2,9), dtype=np.int8)
    npawns[0][material[0][1]] = 1 
    npawns[1][material[1][1]] = 1

    mobtable = mobtable.reshape(2,mobtable.shape[1],8,8)
    mobtable[0] = np.flip(mobtable[0],1)
    othertable = othertable.reshape(2,othertable.shape[1],8,8)
    othertable[0] = np.flip(othertable[0],1)

    interpw = (wkingfile - 1) // 2
    interpb = (bkingfile - 1) // 2

    mobtable[0] = interpw * mobtable[0] + (3-interpw) * np.flip(mobtable[0],2)
    mobtable[1] = interpb * mobtable[1] + (3-interpb) * np.flip(mobtable[1],2)
    othertable[0] = interpw * othertable[0] + (3-interpw) * np.flip(othertable[0],2)
    othertable[1] = interpb * othertable[1] + (3-interpb) * np.flip(othertable[1],2)

    bishoppair = np.zeros(1, dtype=np.int8)

    if material[0,3] == 2:
        bishoppair -= 1
    if material[1,3] == 2:
        bishoppair += 1

    oppCastle = np.zeros(1, dtype=np.int8)
    oppCastle[0] = int((wkingfile >= 5) != (bkingfile >= 5))
    differentCastle = np.zeros(1, dtype=np.int8)
    differentCastle[0] = int((wkingfile >= bkingfile-1) and (bkingfile+1 >= wkingfile))
    #shieldsum = np.zeros(1, dtype=np.int8)
    #shieldsum[0] = np.sum(shield)
    #shieldbasesum = np.zeros(1, dtype=np.int8)
    #shieldbasesum[0] = np.sum(shieldbase)

    #pieceCount = np.zeros(1, dtype=np.int8)
    #pieceCount[0] = np.sum(material[1,2:]) - np.sum(material[0,2:])

    kingPawnEndgame = np.zeros(1, dtype=np.int8)
    kingPawnEndgame[0] = (np.sum(material[1,2:6]) == 0) or (np.sum(material[0,2:6]) == 0)

    rookEndgame = np.zeros(1, dtype=np.int8)
    oppBishopEndgame = np.zeros(1, dtype=np.int8)

    # Endgame handler
    if (np.sum(material[1,2:6]) == 1) and (np.sum(material[0,2:6]) == 1):
        if material[1][4] == 1 and material[0][4] == 1:
            rookEndgame[0] = 1
        if material[1][3] == 1 and material[0][3] == 1 and oppcolor[0][0] == oppcolor[1][1]:
            oppBishopEndgame[0] = 1



    pawnCaptures = captures[:,1,:6]
    tempoCaptures = np.zeros(2, dtype=np.int8)

    if turn:
        tempoCaptures[0] -= np.sum(pawnCaptures[0,2:])
        tempoCaptures[1] += np.sum(pawnCaptures[1,2:])
    else:
        tempoCaptures[0] += np.sum(pawnCaptures[1,2:])
        tempoCaptures[1] -= np.sum(pawnCaptures[0,2:])

    npawns[:,6] = 0

    kingFile = np.zeros((8), dtype=np.int8)
    kingRank = np.zeros((8), dtype=np.int8)

    kingFile[wkingfile-1] += 1
    kingRank[wkingrank-2] += 1
    kingFile[bkingfile-1] -= 1
    kingRank[9-bkingrank] -= 1 
    
    # Create a base case for terms creates more consistent and faster tuning
    kingFile[4] = 0
    kingRank[7] = 0
    passerDistance[4] = 0
    passerSupportDistance[4] = 0


    terms = [
        # Weighting
        [material[0, :] + material[1, :], differentCastle], #differentCastle oppCastle, shieldsum, shieldbasesum
        # Statics
        [material[1] - material[0], sidetomove, pushers, phalanx, chain, baseMob[1]-baseMob[0], passerDistance, passerSupportDistance, bishoppair, kingFile, kingRank],
        # Dynamics
        [mobtable[1].flatten()-mobtable[0].flatten(), ],
        [othertable[1].flatten()-othertable[0].flatten(), ],
        [material[1] - material[0], shield[1]-shield[0], shieldbase[1]-shieldbase[0], kingFile, kingRank, tempoCaptures],
        # Drawishness heuristic
        [npawns[0]],#oppBishopEndgame
        [npawns[1]],#oppBishopEndgame
    ]

    if first:
        startCount = 0
        for group in terms:
            lengths = []
            for item in group:
                startCount += len(item)
                lengths.append(len(item))
            starts.append(startCount)
            sizes.append(lengths)

        #print(starts, sizes)

        print("\nfen " + fen)
        #print(oppBishopEndgame)
        #for a in range(2):
        #    plt.imshow(kwhite[1].reshape((8,8)))
        #    plt.show()
        #    plt.imshow(qwhite[1].reshape((8,8)))
        #    plt.show()
        #    plt.imshow((kwhite+qwhite)[1].reshape((8,8)))
        #    plt.show()
        #    plt.imshow(mobtable[a][6].reshape((8,8)))
        #    plt.show()
        #    break
        #    ...

        #plt.imshow((mobtable[0][8]).reshape((8,8)))
        #plt.show()

        #print(shield)
        #print(shieldbase)
        #sys.exit()

    finalterms = []
    for iterm in range(len(terms)):
        finalterms = finalterms + terms[iterm]

    values = np.concatenate(finalterms, dtype=np.int8) #casting="unsafe")

    inputs.append(values)
    outputs.append(outcome)
    first = False

import torch
from torch.utils.data import DataLoader, TensorDataset

class HCE(torch.nn.Module):
    def __init__(self):
        super().__init__()
        self.npieces = (starts[3] - starts[2]) // 64
        self.nspecial = (starts[4] - starts[3]) // 64

        self.static_terms = torch.nn.Parameter(torch.randn(starts[2]-starts[1]))
        self.dynamic_terms = torch.nn.Parameter(torch.randn(starts[5]-starts[4]))
        self.mobilitytable = torch.nn.Parameter(torch.randn(64))
        self.piecemobility = torch.nn.Parameter(torch.randn(self.npieces))
        self.special = torch.nn.Parameter(torch.randn(self.nspecial))

        self.phase = torch.nn.Parameter(torch.abs(torch.randn(starts[1])))
        self.risk = torch.nn.Parameter(torch.randn(starts[6]-starts[5]))

    def forward(self, x):     
        phase = (x[:,2] +x[:,3] +(x[:,4]*2) +(x[:,5]*4))/24

        mobilityMg = torch.matmul(torch.matmul(x[:, starts[2]:starts[3]].reshape(x.shape[0],self.npieces,64), self.mobilitytable), self.piecemobility) / 3
        mobilityEg = torch.matmul(torch.matmul(x[:, starts[3]:starts[4]].reshape(x.shape[0],self.nspecial,64), self.mobilitytable), self.special) / 3

        statics = mobilityEg + torch.matmul(x[:,starts[1]:starts[2]], self.static_terms)
        dynamics = mobilityMg + torch.matmul(x[:,starts[4]:starts[5]], self.dynamic_terms)
        
        dynamics_weight = torch.clamp(torch.matmul(x[:,starts[0]:starts[1]], self.phase), min=0) #phase

        # TODO: replace phase with self.phase once we get backprop working properly
        score = statics + dynamics * dynamics_weight
        
        pawnscalar = torch.sign(torch.clamp(score, min=0)) * torch.matmul(x[:,starts[6]:starts[7]], self.risk) + torch.sign(torch.clamp(-(score), min=0)) * torch.matmul(x[:,starts[5]:starts[6]], self.risk)
        drawishness_weight = torch.clamp(pawnscalar, min=0)

        finalscore = score / (1 + dynamics_weight + drawishness_weight)

        return torch.tanh(finalscore) 

    def printfinal(self, finalEpoch=False):
        #print((self.risk.detach().numpy()))

        m = 100 / 0.54319 # For tanh this represents the "50%" winning chance

        def printparams(regular, shape, forgrid=True):
            offset = 0
            for size in shape:
                if size >= 64 and forgrid:
                    size = size//64
                a = np.around(regular[offset:offset+size].detach().numpy() * m, decimals=0)
                
                finalstr = "{"
                for i in range(len(a)):
                    finalstr += str(int(a[i])) + ", "
                finalstr += "}"#.4811 +friendly passer king distance

                print(finalstr)
                offset += size
        
        print("\nLinear terms")
        printparams(self.static_terms, sizes[1])
        print("Pt2")
        printparams(self.special, sizes[3])
        print("Pt3")
        printparams(self.dynamic_terms, sizes[4])
        print("\nMobility weights")
        printparams(self.piecemobility, sizes[2])

        print("\nRisk weights")
        print(np.around((self.risk.detach().numpy()) * 1000).astype(np.int32))

        print("\nPhase weights")
        print(np.around((self.phase.detach().numpy()) * 1000).astype(np.int32))

        if finalEpoch or True:
            print("\nBoard Weights")
            print(np.around(self.mobilitytable.detach().numpy().reshape((8,8)) * 1000).astype(np.int32))

        print("===")
        if finalEpoch:
            plt.title("Board")
            plt.imshow(self.mobilitytable.detach().numpy().reshape((8,8)))
            plt.show()


inputs = torch.tensor(np.array(inputs), dtype=torch.float32)
outputs = torch.tensor(outputs, dtype=torch.float32)
size = len(outputs)

# Create a TensorDataset and DataLoader
dataset = TensorDataset(inputs, outputs)
batch_size = 32  # You can adjust the batch size
dataloader = DataLoader(dataset, batch_size=batch_size, shuffle=True)

model = HCE()

criterion = torch.nn.MSELoss(reduction='sum')
optimizer = torch.optim.Adam(model.parameters(), lr=1e-3, weight_decay=1e-4)

epochs = 20

# Training loop
for epoch in range(epochs):  # Adjust the number of epochs
    total = 0
    for batch_x, batch_y in dataloader:
        # Forward pass
        y_pred = model(batch_x)

        # Compute loss
        loss = criterion(y_pred, batch_y)
        total += loss.item()
        
        # Backward pass and optimization
        optimizer.zero_grad()
        loss.backward()
        optimizer.step()

    print(epoch, total/size)
    model.printfinal(epoch == epochs - 1)

# New baseline       .23907
# Only shieldbase    .23996 ~89
# Only shieldnonbase .23977 ~70
# Only shieldfile    .24071 but pawn broken .24004

# Material
# Endgame
# Piece Activity
# King Safety
# Trade pieces when you're up material
# Don't trade off too many pawns (drawishness)
# When down material play for attack/iniative, versa dont allow counterplay when up material
# Spend more time in positions where there are multiple possible moves, i.e. decision nodes
# Should be more or less linear/logistic regression, with easily interpretable values
# The evaluation should represent the practical chances of a position rather than the true value

# 
#.4824 5M ALL
#.4816 5M BISHOP PAIR
#.4850 5M -restricted -pawn structure
#.4842 5M +restricted
#.4818 5M +pawn structure +defended pieces x2
#.4815 5M -defended pieces +unprotected passer x2
#.4820 5M -unprotected passer -king pawn endgame scaling
#.4812. 5M +king pawn endgame scaling +rooks on semi open files
#.4818 -rooks on semi open files +pawn mobility
#.4811 -pawn mobility +blockaded passers
#.4817 -blockaded passer +king rays
#.4793 -king rays +complex attacks
#.4815 -complex attacks +tropism
#.4811 -tropism +pawn tropism
#.4821 -pawn tropism -tempo
#.4803 +tempo +simplish attacks
#.4808 simplish attacks but only in static
#.4814 simplish attacks but only in dynamic
#.4825 -simplish attacks -shieldbase
#.4834 +shieldbase -base mobility
#.4815 +base mobility +enemy passer king distance
#.4811 +friendly passer king distance
#.4814 -passer distance -friendly passser distance, restricted cuts out of base mobility
#.4812 restricted cuts out of main mobility not base
#.4817 restricted cuts out both
#.4805 base mobility counts attacks on friendly squares, restricted cuts only main mobility
#.4810 -shieldbase
#.4816 +shieldbase -both passer king distances
#.4810 +endgame kingfile kingrank
#.4803 both mg and eg kingfile kingrank

# various scaling factors, drawish bishop endgame, king pawn endgame, rook endgame, blocked/rammed pawn position.

# 4755 attacks all
# 4780 no attacking pieces
# 4776 ^ this but modifier to add restricted squares which had piece on them
# 4783 no attacks at all
# 4774 only attacks on pieces


#.4783 current (pawn defended squares are counted if an actual piece is sitting on them)
#.4788 remove middle game king position
#.4773 simple blocked as part of game state
#.4784 -blockaded passer

#.3065 baseline lichessbig
#.3067 no scaling besides phase
#.3084 - passer distance
#.3095 - bishop pair
#.3067 +bp +pd, change passer distance to reflect king in front/behind pawn
#.3062 + pawn captures + different castling scalar + mobility king interpolation
#.3052 - different castling scalar + material in middlegame
#^ produces varying scores ive seen as low as .3039
#.3038 with passers in dynamic (probably not worth the headache imo)
#.3039 tried doing psqt for each piece based on attention map (both mg and eg)

# ideas left to try:
# - scaling dynamics terms in various ways
# - try adding pieces/other stuff to mobtable?, should try kings again since tables r getting flipped now hmmmm
# - friendly king passer escort