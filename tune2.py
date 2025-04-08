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

starts = []
sizes = []
first = True

rays = [ False, False, False, True, True, True, False]
patterns = [ [], [], [N+N+W,N+N+E,S+S+W,S+S+E,W+W+N,W+W+S,E+E+N,E+E+S], [N+W,N+E,S+W,S+E], [N,S,E,W], [N,S,E,W,N+W,N+E,S+W,S+E], [N,S,E,W,N+W,N+E,S+W,S+E]]


for line in tqdm(lines):
    if len(outputs) > 1_000:
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
    mobilities = np.zeros((2,7))

    attackspawn = np.zeros((7))
    attackspiece = np.zeros((7))

    sidetomove = np.zeros(1)

    backwards = np.zeros((2,10))
    passers = np.zeros((2,10))
    passerRank = np.array([
        [0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
        [11,11,11,11,11,11,11,11,11,11],
    ])
    isolated = np.zeros((2,10))


    psqt = np.zeros(64)

    shield = np.zeros((2,10))
    shield2 = np.zeros((2,10))
    shield3 = np.zeros((2,10))

    mobtable = np.zeros((2,7,64))
    kingattacked = np.zeros((2,7))

    kings = [-1, -1]
    kingring = np.zeros((2,120))
    rearpawns = [
        [12,12,12,12,12,12,12,12,12,12],
        [0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    ]

    for i in range(64):
        piece = virtualboard[mailbox[i]]
        piecetype = piece // 2

        if piecetype == 0:
            continue

        piececolor = piece & 1

        material[piececolor, piecetype] += 1

        if piecetype == 1:
            pfile = mailbox[i] % 10
            prank = mailbox[i] // 10
            if piececolor == 0:
                rearpawns[piececolor][pfile] = min(rearpawns[piececolor][pfile], prank)
            else:
                rearpawns[piececolor][pfile] = max(rearpawns[piececolor][pfile], prank)

            # pawn psqt
            j = i
            if piececolor != 1:
                j = i ^ 56

            # Base case for psqt, e3 pawn
            if j != 44:
                psqt[j] += sign[piececolor]

        elif piecetype == 6:
            kings[piececolor] = mailbox[i]
            for off in [N,S,E,W,N+W,N+E,S+W,S+E]:
                kingring[piececolor][mailbox[i] + off] = 1

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

            if piece & 1 == 0:
                if rearpawns[1][pfile - 1] <= prank and rearpawns[1][pfile] <= prank and rearpawns[1][pfile + 1] <= prank:
                    passers[0][pfile] += 1
                    passerRank[0][pfile] = max(prank, passerRank[0][pfile])
                if rearpawns[0][pfile-1] > prank and rearpawns[0][pfile+1] > prank:
                    backwards[0][pfile] += 1
                if rearpawns[0][pfile-1] == 12 and rearpawns[0][pfile+1] == 12:
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
        
        mobility = 0

        for direction in pattern:
            current = sq
            for i in range(1,8):
                current += direction

                if virtualboard[current] == 1:
                    break

                if virtualboard[current] == 0 or ((virtualboard[current] & 1) != (piece & 1)):
                    mobility += 1
                    if kingring[1-(piece & 1)][current] == 1:
                        kingattacked[1-(piece&1)][piecetype] += 1

                    idx = inverse[current]
                    if piece&1 == 1:
                        idx = idx ^ 56

                    mobtable[piece&1][piecetype][idx] += 1

                    # piece attacks
                    if virtualboard[current] > 3:
                        attackspiece[piecetype] += sign[piece&1]
                    elif virtualboard[current] > 1:
                        attackspawn[piecetype] += sign[piece&1]

                if virtualboard[current] != 0 or (not isray):
                    break


        mobilities[piece & 1][piecetype] += mobility

    for x in [-1,0,1]:
        wspawn = rearpawns[1][wkingfile + x]
        if wspawn > 0:
            shield[1][wkingfile + x] = 1
        if wspawn == 8:
            shield2[1][wkingfile + x] = 1
        if wspawn > 0 and wspawn < wkingrank:
            shield3[1][wkingfile + x] = 1

        bspawn = rearpawns[0][bkingfile + x]
        if bspawn < 10:
            shield[0][bkingfile + x] = 1
        if bspawn == 3:
            shield2[0][bkingfile + x] = 1
        if bspawn < 10 and bspawn > bkingrank:
            shield3[0][bkingfile + x] = 1

    shield = 1-shield
    shield2 = 1-shield2
    shield3 = 1-shield3

    # Clipping passers
    passers = np.clip(passers, 0, 1) # We dont count doubled pawns as multiple passers
    pushers = np.zeros(10)


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

    imbalance = np.dot(material[1]-material[0], np.array([0,1,3,3,5,9,0]))

    downmaterial = [max(0,imbalance), max(0,-imbalance)]

    terms = [
        [material[0, :] + material[1, :]],
        [material[1, :] - material[0, :], psqt, isolated[1] - isolated[0], passers[1] - passers[0], backwards[1] - backwards[0], shield[1] - shield[0], sidetomove, attackspiece, attackspawn, pushers],
        [mobtable.flatten(),],
        [np.array([1,]), [downmaterial[1],], shield[0], kingattacked[0], np.array([1,]), [downmaterial[0],], shield[1], kingattacked[1]],
    ]

    if first:
        names = ["phase", "mobility", "kingsafety", "linear"]

        startCount = 0
        for group in terms:
            lengths = []
            starts.append(startCount)
            for item in group:
                startCount += len(item)
                lengths.append(len(item))
            sizes.append(lengths)

        print(starts, sizes)

        print("\nfen " + fen)
        #print(pushers)
        #print(imbalance, downmaterial)
        for a in range(2):
            for b in range(1,7):
                #plt.imshow(mobtable[a][b].reshape((8,8)))
                #plt.show()
                ...
        #sys.exit()

    for iterm in range(len(terms)):
        terms[iterm] = np.concatenate(terms[iterm])

    values = np.concatenate(terms)

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

        self.piecemobility = torch.nn.Parameter(torch.randn(7,1))
        self.taperpiecemobility = torch.nn.Parameter(torch.randn(7,1))

        self.danger = torch.nn.Parameter(torch.randn(19))
        self.taperdanger = torch.nn.Parameter(torch.randn(19))

        print("postmul",torch.mul(self.mobilitytable, self.piecemobility).shape)


    def forward(self, x):
        phase = (x[:,2] +x[:,3] +(x[:,4]*2) +(x[:,5]*4))/24

        mob = x[:,starts[2]:starts[3]]

        kinghalfs = x[:,starts[3]:]

        bkinghalf = kinghalfs[:,:kinghalfs.shape[1]//2]
        wkinghalf = kinghalfs[:,kinghalfs.shape[1]//2:]

        torch.matmul(bkinghalf, self.danger)
        
        wdanger = (torch.matmul(wkinghalf, self.danger) * phase) + (torch.matmul(wkinghalf, self.taperdanger) * (1-phase))
        bdanger = (torch.matmul(wkinghalf, self.danger) * phase) +  (torch.matmul(wkinghalf, self.taperdanger) * (1-phase))

        wdanger = torch.clamp(wdanger, min=0)
        bdanger = torch.clamp(bdanger, min=0)

        bmob = torch.matmul(mob[:,:mob.shape[1]//2], torch.mul(self.mobilitytable, self.piecemobility).flatten())
        wmob = torch.matmul(mob[:,mob.shape[1]//2:], torch.mul(self.mobilitytable, self.piecemobility).flatten())

        bmob2 = torch.matmul(mob[:,:mob.shape[1]//2], torch.mul(self.tapermobilitytable, self.taperpiecemobility).flatten())
        wmob2 = torch.matmul(mob[:,mob.shape[1]//2:], torch.mul(self.tapermobilitytable, self.taperpiecemobility).flatten())

        score = torch.matmul(x[:,starts[1]:starts[2]], self.terms)
        score2 = torch.matmul(x[:,starts[1]:starts[2]], self.taperterms)

        return torch.tanh(((score+wmob-bmob) * phase) + ((score2+wmob2-bmob2) * (1-phase))) + bdanger - wdanger

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