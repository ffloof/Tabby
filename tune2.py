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


for line in tqdm(lines):
    if len(outputs) > 2_000_000:
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


    phalanxOpen = np.zeros((2, 1), dtype=np.int8)
    phalanxClosed = np.zeros((2, 1), dtype=np.int8)
    chainOpen = np.zeros((2, 1), dtype=np.int8)
    chainClosed = np.zeros((2, 1), dtype=np.int8) 


    passerDistance = np.zeros((2,8), dtype=np.int8)
    passerRank = np.array([
        [0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
        [11,11,11,11,11,11,11,11,11,11],
    ], dtype=np.int8)

    shield = np.zeros((2,10), dtype=np.int8)

    mobtable = np.zeros((2,11,64), dtype=np.int8)

    kings = [-1, -1]
    rearpawns = [
        [11,11,11,11,11,11,11,11,11,11],
        [0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    ]
    
    captures = np.zeros((2,7,7), dtype=np.int8)
    restricted = np.zeros((2,7), dtype=np.int8)
    
    for i in range(64):
        piece = virtualboard[mailbox[i]]
        piecetype = piece // 2

        if piecetype == 0:
            continue

        piececolor = piece & 1

        material[piececolor, piecetype] += 1

        if piecetype > 0:
            mobtable[piece&1][piecetype][i] += 1

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

    for sq in range(len(virtualboard)):
        piece = virtualboard[sq]
        piecetype = piece // 2
        isray = rays[piecetype]
        pattern = patterns[piecetype]

        if piecetype == 1:
            pfile = sq % 10
            prank = sq // 10

            j = inverse[sq]

            if piece & 1 == 0:
                if virtualboard[sq + W] == 2 or virtualboard[sq + E] == 2:
                    if rearpawns[1][pfile] == 0:
                        #mobtable[0][7][inverse[sq]] = 1
                        phalanxOpen[0] += 1
                    else:
                        #mobtable[0][8][inverse[sq]] = 1
                        phalanxClosed[0] += 1

                if virtualboard[sq + W + N] == 2 or virtualboard[sq + E + N] == 2:
                    if rearpawns[1][pfile] == 0:
                        #mobtable[0][9][inverse[sq]] = 1
                        chainOpen[0] += 1
                    else:
                        #mobtable[0][10][inverse[sq]] = 1
                        chainClosed[0] += 1

                if rearpawns[1][pfile - 1] <= prank and rearpawns[1][pfile] <= prank and rearpawns[1][pfile + 1] <= prank:
                    passerRank[0][pfile] = max(prank, passerRank[0][pfile])

            else:
                if virtualboard[sq + W] == 3 or virtualboard[sq + E] == 3:
                    if rearpawns[0][pfile] == 11:
                        #mobtable[1][7][inverse[sq]] = 1
                        phalanxOpen[1] += 1
                    else:
                        #mobtable[1][8][inverse[sq]] = 1
                        phalanxClosed[1] += 1

                if virtualboard[sq + W + S] == 3 or virtualboard[sq + E + S] == 3:
                    if rearpawns[0][pfile] == 11:
                        #mobtable[1][9][inverse[sq]] = 1
                        chainOpen[1] += 1
                    else:
                        #mobtable[1][10][inverse[sq]] = 1
                        chainClosed[1] += 1


                if rearpawns[0][pfile - 1] >= prank and rearpawns[0][pfile] >= prank and rearpawns[0][pfile + 1] >= prank:
                    passerRank[1][pfile] = min(prank, passerRank[1][pfile])
        
        if piece == 2:
            pattern = [S+W, S+E]
        if piece == 3:
            pattern = [N+W, N+E]
        

        for direction in pattern:
            current = sq
            for i in range(1,8):
                current += direction

                if virtualboard[current] == 1:
                    break

                if virtualboard[current] == 0 or ((virtualboard[current] & 1) != (piece & 1)):
                    pawnDefence = False
                    if piecetype != 1:
                        mobtable[piece&1][piecetype][inverse[current]] += 1

                        if piece & 1 == 0 and (virtualboard[current+S+W] == 3 or virtualboard[current+S+E] == 3):
                            pawnDefence = True
                        elif piece & 1 == 1 and (virtualboard[current+N+W] == 2 or virtualboard[current+N+E] == 2):
                            pawnDefence = True

                    if pawnDefence:
                        restricted[piece & 1][piecetype] += 1

                    if virtualboard[current] != 0 and ((virtualboard[current] & 1) != (piece & 1)):
                        captures[piece & 1][piecetype][virtualboard[current] // 2] += 1



                if virtualboard[current] != 0 or (not isray):
                    break

    for x in [-1,0,1]:
        wspawn = rearpawns[1][wkingfile + x]
        if wspawn > 0:
            shield[1][wkingfile + x] = 1

        bspawn = rearpawns[0][bkingfile + x]
        if bspawn < 10:
            shield[0][bkingfile + x] = 1

    pushers = np.zeros(8, dtype=np.int8)


    for i in range(10):
        wPass = 11 - passerRank[1][i]
        bPass = passerRank[0][i]
        wPass -= 2
        bPass -= 2

        if wPass >= 0:
            pushers[wPass] += 1
            passerDistance[1][abs(i-bkingfile)] += 1

        if bPass >= 0:
            pushers[bPass] -= 1
            passerDistance[0][abs(i-wkingfile)] += 1

    sidetomove[0] = sign[turn]

    npawns = np.zeros((2,9), dtype=np.int8)
    npawns[0][material[0][1]] = 1 
    npawns[1][material[1][1]] = 1


    passerDistance[:,0] = 0

    mobtable = mobtable.reshape(2,mobtable.shape[1],8,8)
    mobtable[0] = np.flip(mobtable[0],1)

    if wkingfile < 5:
        mobtable[1,:,4:8] = np.flip(mobtable[1,:,4:8], 2)
        mobtable[0,:,0:4] = np.flip(mobtable[0,:,0:4], 2)

    if bkingfile < 5:
        mobtable[0,:,4:8] = np.flip(mobtable[0,:,4:8], 2)
        mobtable[1,:,0:4] = np.flip(mobtable[1,:,0:4], 2)

    captures[:,:,6] = 0 

    bishoppair = np.zeros(1, dtype=np.int8)

    if material[0,3] == 2:
        bishoppair -= 1
    if material[1,3] == 2:
        bishoppair += 1

    terms = [
        [material[0, :] + material[1, :], sidetomove],
        [material[1, :] - material[0, :], pushers, shield[1]-shield[0], sidetomove, passerDistance[1] - passerDistance[0], restricted[1] - restricted[0], (captures[1] - captures[0]).flatten(), bishoppair, phalanxOpen[1]-phalanxOpen[0], phalanxClosed[1] - phalanxClosed[0], chainOpen[1] - chainOpen[0], chainClosed[1] - chainClosed[0] ], 
        [mobtable[0].flatten()],
        [mobtable[1].flatten()],
        [npawns[0]],
        [npawns[1]],
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
        print(chainOpen)
        print(chainClosed)
        #print(bishoppair)
        #for a in range(2):
        #    plt.imshow(mobtable[a][1].reshape((8,8)))
        #    plt.show()
        #    ...

        #plt.imshow((mobtable[0][8]).reshape((8,8)))
        #plt.show()

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

        self.terms = torch.nn.Parameter(torch.randn(starts[2]-starts[1]))
        self.taperterms = torch.nn.Parameter(torch.randn(starts[2]-starts[1]))

        self.mobilitytable = torch.nn.Parameter(torch.randn(64))
        self.tapermobilitytable = torch.nn.Parameter(torch.randn(64))

        self.npieces = (starts[3] - starts[2]) // 64
        self.piecemobility = torch.nn.Parameter(torch.randn(self.npieces))
        self.taperpiecemobility = torch.nn.Parameter(torch.randn(self.npieces))

        self.risk = torch.nn.Parameter(torch.randn(starts[5]-starts[4]))

        self.tempomulter = torch.nn.Parameter(torch.randn(1))

    def forward(self, x):
        phase = (x[:,2] +x[:,3] +(x[:,4]*2) +(x[:,5]*4))/24

        blackattention = torch.matmul(x[:, starts[2]:starts[3]].reshape(x.shape[0],self.npieces,64), self.mobilitytable)
        whiteattention = torch.matmul(x[:, starts[3]:starts[4]].reshape(x.shape[0],self.npieces,64), self.mobilitytable)

        netmobility = torch.matmul(whiteattention, self.piecemobility) - torch.matmul(blackattention, self.piecemobility)
        netmobility2 = torch.matmul(whiteattention, self.taperpiecemobility) - torch.matmul(blackattention, self.taperpiecemobility)

        score = torch.matmul(x[:,starts[1]:starts[2]], self.terms)
        score2 = torch.matmul(x[:,starts[1]:starts[2]], self.taperterms)

        bonus = torch.abs(netmobility) *  x[:,7] * self.tempomulter
        bonus2 = torch.abs(netmobility2) * x[:,7] * self.tempomulter

        midscore = ((score + netmobility + bonus) * phase)
        endscore = ((score2 + netmobility2 + bonus2) * (1-phase))

        finalscore = midscore + endscore
        scalew = torch.clamp(finalscore, min=0) * torch.matmul(x[:,starts[5]:starts[6]], self.risk)
        scaleb = torch.clamp(-(finalscore), min=0) * torch.matmul(x[:,starts[4]:starts[5]], self.risk)

        return torch.tanh(finalscore + (scalew - scaleb)) 

    def printfinal(self, finalEpoch=False):
        print((self.risk.detach().numpy()))

        print("TempoMulter", self.tempomulter)


        m = 100 / 0.54319 # For tanh this represents the "50%" winning chance
        riskNormalizer = self.risk.detach().numpy()[8] + 1

        def printparams(regular, tapered, shape, forgrid=True):
            offset = 0
            for size in shape:
                if size >= 64 and forgrid:
                    size = size//64
                a = np.around(regular[offset:offset+size].detach().numpy() * m * riskNormalizer, decimals=0)
                b = np.around(tapered[offset:offset+size].detach().numpy() * m * riskNormalizer , decimals=0)
                
                finalstr = "{"
                for i in range(len(a)):
                    finalstr += "T(" + str(int(a[i])) + "," + str(int(b[i])) + "), "
                finalstr += "}"
                print(finalstr)
                offset += size
        
        print("\nLinear terms")
        printparams(self.terms, self.taperterms, sizes[1])

        print("\nMobility weights")
        printparams(self.piecemobility, self.taperpiecemobility, sizes[2])

        print("\nRisk weights")
        print(np.around((1+self.risk.detach().numpy())/riskNormalizer , decimals=3))

        if finalEpoch:
            print("\nBoard Weights")
            print(np.around(self.mobilitytable.detach().numpy().reshape((8,8)), decimals=3))

        print("===")
        if finalEpoch:
            plt.title("Board")
            plt.imshow(self.mobilitytable.detach().numpy().reshape((8,8)))
            plt.show()






inputs = torch.FloatTensor(np.array(inputs))
outputs = torch.FloatTensor(outputs)
size = len(outputs)

# Create a TensorDataset and DataLoader
dataset = TensorDataset(inputs, outputs)
batch_size = 32  # You can adjust the batch size
dataloader = DataLoader(dataset, batch_size=batch_size, shuffle=True)

model = HCE()

criterion = torch.nn.MSELoss(reduction='sum')
optimizer = torch.optim.Adam(model.parameters(), lr=1e-3, weight_decay=1e-5)

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


# tapered and pawn scaled 2M
# material .3456
# + weighted mobility .3207
# + passerRank 0.3164
# + shield .3166
# + tempo  .3139
# + passerKingFileDistance .3115
# + isolated .3105
# + backwards .3092
# + open distinction .3089
# + restricted .3066
# + attacks .3042
# + bishoppair .3035
# - backwards and isolated
# + phalanx and chain .3027