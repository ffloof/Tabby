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

sizes = []
net_size = 0
linear_start = 0
linear_size = 0
first = True

rays = [ False, False, False, True, True, True, False]
patterns = [ [], [], [N+N+W,N+N+E,S+S+W,S+S+E,W+W+N,W+W+S,E+E+N,E+E+S], [N+W,N+E,S+W,S+E], [N,S,E,W], [N,S,E,W,N+W,N+E,S+W,S+E], [N,S,E,W,N+W,N+E,S+W,S+E]]

def pawncheck(virtualboard, index, pawntype):
    if pawntype == 3:
        return virtualboard[index + S + W] == pawntype or virtualboard[index + S + E] == pawntype
    if pawntype == 2:
        return virtualboard[index + N + W] == pawntype or virtualboard[index + N + E] == pawntype

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

    material = np.zeros((2, 7))
    psqt = np.zeros((7, 64))
    mobilities = np.zeros(7)
    restrictedmobilities = np.zeros(7)

    attacks_pawn = np.zeros(7)
    attacks_other = np.zeros(7)

    backwardsc = np.zeros(8)
    backwardso = np.zeros(8)
    passers = np.zeros(8)

    kingopen = np.zeros(2)
    sidetomove = np.zeros(1)

    npawns = np.zeros(2)

    kings = [-1, -1]


    for i in range(64):
        piece = virtualboard[mailbox[i]]
        piecetype = piece // 2

        if piecetype == 0:
            continue

        piececolor = piece & 1

        material[piececolor, piecetype] += 1

        j = i
        if piececolor != 1:
            j = i ^ 56

        psqt[piecetype][j] += sign[piececolor]

        if piecetype == 6:
            kings[piececolor] = mailbox[i]

        if piecetype == 1:
            npawns[piece & 1] += 1

    
    wshield = []
    if kings[1] % 10 < 5:
        wshield = [33,32,31]
    else:
        wshield = [36,37,38]

    bshield = []
    if kings[0] % 10 < 5:
        bshield = [83,82,81]
    else:
        bshield = [86,87,88]
    
    pawnshield = np.zeros((2,3))

    for i in range(len(bshield)):
        if virtualboard[bshield[i]] != 2 and virtualboard[bshield[i] + N] != 2:
            pawnshield[0,i] += 1

    for i in range(len(wshield)):
        if virtualboard[wshield[i]] != 3 and virtualboard[wshield[i] + S] != 3:
            pawnshield[1,i] += 1

    inCheck = np.zeros(2)
    inCheck[0] = int(pawncheck(virtualboard, kings[0], 3))
    inCheck[1] = int(pawncheck(virtualboard, kings[0], 2))

    for sq in range(len(virtualboard)):
        piece = virtualboard[sq]
        piecetype = piece // 2
        isray = rays[piecetype]
        pattern = patterns[piecetype]

        if piecetype == 1:
            continue
        
        mobility = 0
        restrictedmobility = 0
        attackspawn = 0
        attacksother = 0

        for direction in pattern:
            current = sq
            for i in range(1,8):
                current += direction

                if virtualboard[current] == 1:
                    break

                if virtualboard[current] == 0 or ((virtualboard[current] & 1) != (piece & 1)):
                    #mobility += 1
                    j = inverse[current]
                    
                    if piece & 1 == 0:
                        j ^= 56
                    
                    # is restricted?

                    isRestricted = False
                    if piece & 1 == 1:
                        isRestricted = pawncheck(virtualboard, current, 2)
                    else:
                        isRestricted = pawncheck(virtualboard, current, 3)

                    mobility += 1                    
                    if isRestricted:
                        restrictedmobility += 1


                    # piece attacks
                    if virtualboard[current] > 3:
                        attacksother += 1
                        if virtualboard[current] >= 12:
                            inCheck[piece & 1] = 1
                    elif virtualboard[current] > 1:
                        attackspawn += 1


                if virtualboard[current] != 0 or (not isray):
                    break
        
        #if virtualboard[sq] > 1:
            #print("  pPnNbBrRqQkK"[virtualboard[sq]] , domesticmobility * sign[piece & 1], foreignmobility * sign[piece & 1])
        mobilities[piecetype] += sign[piece & 1] * mobility
        restrictedmobilities[piecetype] += sign[piece & 1] * restrictedmobility
        attacks_pawn[piecetype] += sign[piece & 1] * attackspawn
        attacks_other[piecetype] += sign[piece & 1] * attacksother

    for side in range(2):
        if inCheck[side] == 1:
            continue
        start = kings[side]
        hit = 0
        for direction in patterns[6]:
            i = start + direction
            while virtualboard[i] == 0:
                i += direction
                hit += 1
        kingopen[int(side)] += hit

    sidetomove[0] = sign[turn]

    manualterms = [material[0, :], material[1, :]]
    linearterms = [material[1, :] - material[0, :], psqt[1, :], sidetomove, mobilities]

    if first:
        for item in linearterms:
            sizes.append(item.shape[0])

        print("\nfen " + fen)
        print(material)
        #print(mobilities)
        print(inCheck)
        #print("domestic",domesticmobilities)
        #print("foreign",foreignmobilities)
        #print("restricted",restrictedmobilities)
        #print("kingvmob", kingopen)
        #print("backwardsc", backwardsc)
        #print("backwardso", backwardso)
        #print("passers", passers)
        #print("attacks pawn", attacks_pawn)
        #print("attacks other", attacks_other)
        #print("defends pawn", defends_pawn)
        #print("print other", defends_other)
        #print(kingopen)
        print("npawns", npawns)
        #sys.exit()

    manualterms = np.concatenate(manualterms)
    linearterms = np.concatenate(linearterms)

    if first:
        linear_start = manualterms.shape[0]
        linear_size = linearterms.shape[0]
        print(linear_start, linear_size)

    values = np.concatenate([manualterms, linearterms])

    inputs.append(values)
    outputs.append(outcome)
    first = False

import torch
from torch.utils.data import DataLoader, TensorDataset

class HCE(torch.nn.Module):
    def __init__(self):
        super().__init__()

        self.terms = torch.nn.Parameter(torch.randn(linear_size))
        self.pawnscale = torch.nn.Parameter(torch.randn(1))

    def forward(self, x):
        pawnsw = x[:,1+7]
        pawnsb = x[:,1]

        knights = x[:,2] + x[:,2+7]
        bishops = x[:,3] + x[:,3+7]
        rooks = x[:,4] + x[:,4+7]
        queens = x[:,5] + x[:,5+7]

        score = torch.matmul(x[:,linear_start:], self.terms)

        return torch.tanh(score)

    def printfinal(self, finalEpoch=False):
        m = 100 / 0.54319 # For tanh this represents the "50%" winning chance
        print ("m=",m)
        offset = 0
        for size in sizes:
            #if size == 64:
            #    if finalEpoch:
            #        plt.imshow((self.midgame[offset:offset+size].detach().numpy() * m).astype(np.int32).reshape((8,8)))
            #        plt.show()
            #        plt.imshow((self.endgame[offset:offset+size].detach().numpy() * m).astype(np.int32).reshape((8,8)))
            #        plt.show()
            print((self.terms[offset:offset+size].detach().numpy() * m).astype(np.int32))
            offset += size

        print("===")


x = torch.FloatTensor(np.array(inputs))
y = torch.FloatTensor(outputs)
size = len(outputs)

# Create a TensorDataset and DataLoader
dataset = TensorDataset(x, y)
batch_size = 32  # You can adjust the batch size
dataloader = DataLoader(dataset, batch_size=batch_size, shuffle=True)

model = HCE()

criterion = torch.nn.MSELoss(reduction='sum')
optimizer = torch.optim.Adam(model.parameters(), lr=1e-3, weight_decay=1e-5)

epochs = 12

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

# lots of tapering
# material + pawns + tempo = 0.2671
# + mobilities = 0.2608
# + double taper = 0.2587
# + kingopen + incheck = 0.2562
