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
    if len(outputs) > 1_000_000:
        break

    packed = line.split("c9")
    fen = packed[0].strip()
    outcome = outcomescore[packed[1].strip().replace(";", "").replace("\"", "")]

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

    backwards = np.zeros((2,10), dtype=np.int8)
    passers = np.zeros((2,10), dtype=np.int8)
    passerRank = np.array([
        [0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
        [11,11,11,11,11,11,11,11,11,11],
    ], dtype=np.int8)

    isolated = np.zeros((2,10), dtype=np.int8)

    shield = np.zeros((2,10), dtype=np.int8)

    mobtable = np.zeros((2,7,64), dtype=np.int8)

    kings = [-1, -1]
    rearpawns = [
        [11,11,11,11,11,11,11,11,11,11],
        [0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    ]
    
    pawnAttacksMap = np.zeros((2,64), dtype=np.int8)
    kingRingMap = np.zeros((2,64), dtype=np.int8)
    standers = np.zeros((2,7,64), dtype=np.int8)
    
    for i in range(64):
        piece = virtualboard[mailbox[i]]
        piecetype = piece // 2

        if piecetype == 0:
            continue

        piececolor = piece & 1

        material[piececolor, piecetype] += 1

        if piecetype > 0:
            standers[piece&1][piecetype][i] = 1

        if piecetype == 1:
            pfile = mailbox[i] % 10
            prank = mailbox[i] // 10
            if piececolor == 0:
                rearpawns[piececolor][pfile] = min(rearpawns[piececolor][pfile], prank)
            else:
                rearpawns[piececolor][pfile] = max(rearpawns[piececolor][pfile], prank)

        elif piecetype == 6:
            kings[piececolor] = mailbox[i]
            for off in [N,S,E,W,N+W,N+E,S+W,S+E]:
                spot = inverse[mailbox[i] + off]
                if spot != -1:
                    kingRingMap[piececolor][spot] = 1

    wkingfile = kings[1] % 10
    wkingrank = kings[1] // 10
    bkingfile = kings[0] % 10
    bkingrank = kings[0] // 10
    

    for sq in range(len(virtualboard)):
        piece = virtualboard[sq]
        piecetype = piece // 2
        isray = rays[piecetype]
        pattern = patterns[piecetype]

        if piece == 2:
            pattern = [S+W,S+E]
        elif piece == 3:
            pattern = [N+W, N+E]

        if piecetype == 1:
            pfile = sq % 10
            prank = sq // 10

            j = inverse[sq]

            if piece & 1 == 0:
                if rearpawns[1][pfile - 1] <= prank and rearpawns[1][pfile] <= prank and rearpawns[1][pfile + 1] <= prank:
                    passers[0][pfile] += 1
                    passerRank[0][pfile] = max(prank, passerRank[0][pfile])
                if rearpawns[0][pfile-1] > prank and rearpawns[0][pfile+1] > prank:
                    backwards[0][pfile] += 1
                if rearpawns[0][pfile-1] == 11 and rearpawns[0][pfile+1] == 11:
                    isolated[0][pfile] += 1

            else:
                if rearpawns[0][pfile - 1] >= prank and rearpawns[0][pfile] >= prank and rearpawns[0][pfile + 1] >= prank:
                    passers[1][pfile] += 1
                    passerRank[1][pfile] = min(prank, passerRank[1][pfile])
                if rearpawns[1][pfile-1] < prank and rearpawns[1][pfile+1] < prank:
                    backwards[1][pfile] += 1
                if rearpawns[1][pfile-1] == 0 and rearpawns[1][pfile+1] == 0:
                    isolated[1][pfile] += 1

            #continue
        

        for direction in pattern:
            current = sq
            for i in range(1,8):
                current += direction

                if virtualboard[current] == 1:
                    break
                
                if piecetype == 1:
                    pawnAttacksMap[piece & 1][inverse[current]] += 1

                if virtualboard[current] == 0 or ((virtualboard[current] & 1) != (piece & 1)) or piecetype == 1:
                    mobtable[piece&1][piecetype][inverse[current]] += 1

                if virtualboard[current] != 0 or (not isray):
                    break

    for x in [-1,0,1]:
        wspawn = rearpawns[1][wkingfile + x]
        if wspawn > 0:
            shield[1][wkingfile + x] = 1

        bspawn = rearpawns[0][bkingfile + x]
        if bspawn < 10:
            shield[0][bkingfile + x] = 1


    # Clipping passers
    passers = np.clip(passers, 0, 1) # We dont count doubled pawns as multiple passers
    pushers = np.zeros(10, dtype=np.int8)


    for i in range(10):
        wPass = 11 - passerRank[1][i]
        bPass = passerRank[0][i]
        wPass -= 3
        bPass -= 3

        if wPass >= 0:
            pushers[wPass] += 1
        if bPass >= 0:
            pushers[bPass] -= 1


    if bkingfile < 5:
        isolated[0] = isolated[0,::-1]
        backwards[0] = backwards[0,::-1]
        passers[1] = passers[1,::-1]
   
    if wkingfile < 5:
        passers[0] = passers[0,::-1]
        isolated[1] = isolated[1,::-1]
        backwards[1] = backwards[1,::-1]

    sidetomove[0] = sign[turn]

    npawns = np.zeros((2,9), dtype=np.int8)
    npawns[0][material[0][1]] = 1 
    npawns[1][material[1][1]] = 1

    terms = [
        [material[0, :] + material[1, :]],
        [material[1, :] - material[0, :], isolated[1] - isolated[0], passers[1] - passers[0], backwards[1] - backwards[0], shield[1] - shield[0], sidetomove, pushers, ],
        [mobtable[0].flatten(), standers[0].flatten()],
        [mobtable[1].flatten(), standers[1].flatten()],
        [pawnAttacksMap[1], standers[1].flatten(), kingRingMap[1]],
        [pawnAttacksMap[0], standers[0].flatten(), kingRingMap[0]],
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

        print(starts, sizes)

        print("\nfen " + fen)
        #for a in range(2):
        #    plt.imshow(kingRingMap[a].reshape((8,8)))
        #    plt.show()
        #    ...

        #plt.imshow((mobtable[0][4] + pawnAttacksMap[1]).reshape((8,8)))
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

        self.mobilitytable = torch.nn.Parameter(torch.randn(1,64))
        self.tapermobilitytable = torch.nn.Parameter(torch.randn(1,64))

        self.npieces = (starts[3] - starts[2]) // 64
        self.piecemobility = torch.nn.Parameter(torch.randn(self.npieces))
        self.taperpiecemobility = torch.nn.Parameter(torch.randn(self.npieces))

        self.nsquares = (starts[5]-starts[4])//64

        self.gridweights = torch.nn.Parameter(torch.randn(self.nsquares))
        self.tapergridweights = torch.nn.Parameter(torch.randn(self.nsquares))

        self.risk = torch.nn.Parameter(torch.randn(starts[7]-starts[6]))
        self.taperrisk = torch.nn.Parameter(torch.randn(starts[7]-starts[6]))


    def forward(self, x):
        phase = (x[:,2] +x[:,3] +(x[:,4]*2) +(x[:,5]*4))/24

        mob = x[:,starts[2]:starts[3]]

        material = (torch.matmul(x[:,starts[1]:starts[1]+7], self.terms[0:7])* phase) + (torch.matmul(x[:,starts[1]:starts[1]+7], self.taperterms[0:7])* (1-phase))

        wup = torch.clamp(material, min=0)
        bup = torch.clamp(-material, min=0)


        normal = self.mobilitytable
        normalTaper = self.tapermobilitytable

        scaleb = bup * ((torch.matmul(x[:,starts[6]:starts[7]], self.risk)*phase) + (torch.matmul(x[:,starts[6]:starts[7]], self.taperrisk)*(1-phase)))
        scalew = wup * ((torch.matmul(x[:,starts[7]:starts[8]], self.risk)*phase) + (torch.matmul(x[:,starts[7]:starts[8]], self.taperrisk)*(1-phase)))

        inverse = torch.flip(normal.reshape((1,8,8)), [1,]).reshape((1,64))
        inverseTaper = torch.flip(normalTaper.reshape((1,8,8)), [1,]).reshape((1,64))

        bgrids = x[:,starts[4]:starts[5]].reshape(x.shape[0], self.nsquares, 64).movedim(1,2)
        wgrids = x[:,starts[5]:starts[6]].reshape(x.shape[0], self.nsquares, 64).movedim(1,2)

        wpost = torch.matmul(wgrids,self.gridweights) + normal
        bpost = torch.matmul(bgrids,self.gridweights) + inverse
        wtaperpost = torch.matmul(wgrids,self.tapergridweights) + normalTaper
        btaperpost = torch.matmul(bgrids,self.tapergridweights) + inverseTaper

        blackmob = torch.matmul(x[:, starts[2]:starts[3]].reshape(x.shape[0],self.npieces,64).movedim(1,2), self.piecemobility)
        whitemob = torch.matmul(x[:, starts[3]:starts[4]].reshape(x.shape[0],self.npieces,64).movedim(1,2), self.piecemobility)
        blackmobtaper = torch.matmul(x[:, starts[2]:starts[3]].reshape(x.shape[0],self.npieces,64).movedim(1,2), self.taperpiecemobility)
        whitemobtaper = torch.matmul(x[:, starts[3]:starts[4]].reshape(x.shape[0],self.npieces,64).movedim(1,2), self.taperpiecemobility)

        netmob = (whitemob * wpost).sum(dim=1) - (blackmob * bpost).sum(dim=1)
        netmob2 = (whitemobtaper * wtaperpost).sum(dim=1) - (blackmobtaper * btaperpost).sum(dim=1)

        score = torch.matmul(x[:,starts[1]:starts[2]], self.terms)
        score2 = torch.matmul(x[:,starts[1]:starts[2]], self.taperterms)

        return torch.tanh(((score + netmob) * phase) + ((score2 + netmob2) * (1-phase)) + (scalew - scaleb)) 

    def printfinal(self, finalEpoch=False):
        m = 100 / 0.54319 # For tanh this represents the "50%" winning chance
        '''
        offset = 0
        for i in range len(starts):
            print("\n", names[i])

            offset = 0
            for size in sizes[i]:



            if size == 64:
                if finalEpoch:
                    plt.imshow((self.terms[offset:offset+size].detach().numpy() * m).astype(np.int32).reshape((8,8)))
                    plt.show()
                    plt.imshow((self.taperterms[offset:offset+size].detach().numpy() * m).astype(np.int32).reshape((8,8)))
                    plt.show()

                offset += size

            print((self.terms[offset:offset+size].detach().numpy() * m).astype(np.int32))
            print((self.taperterms[offset:offset+size].detach().numpy() * m).astype(np.int32))

            print("")
            offset += size

        print("mobility")

        knight = (m * self.mobilitytable * self.piecemobility[2]).detach().numpy()
        print(knight)

        taperknight = (m * self.tapermobilitytable * self.piecemobility[2]).detach().numpy()
        print(taperknight)

        print(self.piecemobility / self.piecemobility[2])
        print(self.taperpiecemobility / self.piecemobility[2])
        if finalEpoch:
            plt.imshow(knight.astype(np.int32).reshape((8,8)))
            plt.show()
            plt.imshow(taperknight.astype(np.int32).reshape((8,8)))
            plt.show() '''

        print("Piece weights")
        print(np.around(self.piecemobility.detach().numpy(), decimals=4))
        print(np.around(self.taperpiecemobility.detach().numpy(), decimals=4))

        print("Grid weights")
        print(np.around(self.gridweights.detach().numpy(), decimals=4))
        print(np.around(self.tapergridweights.detach().numpy(), decimals=4))

        print("Risk weights")
        print(np.around(self.risk.detach().numpy(), decimals=4))
        print(np.around(self.taperrisk.detach().numpy(), decimals=4))

        if finalEpoch:
            plt.title("Middlegame")
            plt.imshow(self.mobilitytable.detach().numpy().reshape((8,8)))
            plt.show()
            plt.title("Endgame")
            plt.imshow(self.tapermobilitytable.detach().numpy().reshape((8,8)))
            plt.show()


        print("===")


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


# no tapering
# material + pawns + tempo + mobilities = 0.2706
# + isolated + backwards + passers = 0.2647
# + shield * queens = 0.2641
# isolated and passer inversion = .2639
# + backwards and race = 0.259


# lots of tapering
# material + pawns + tempo = 0.2671
# + mobilities = 0.2608
# + shield * queens = 0.2539

# current 0.2533
# current 0.2497
# current 0.2491 / 0.2487
# current 0.2468
# 0.2450
# 0.2398