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
    "0-1":0,
    "1/2-1/2":0.5,
}

sign = [-1,1]
phase_level = [0,0,2,2,3,8,0] # sum 44

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

sizes = []
first = True

rays = [ False, False, False, True, True, True, False]
patterns = [ [], [], [N+N+W,N+N+E,S+S+W,S+S+E,W+W+N,W+W+S,E+E+N,E+E+S], [N+W,N+E,S+W,S+E], [N,S,E,W], [N,S,E,W,N+W,N+E,S+W,S+E], [N,S,E,W,N+W,N+E,S+W,S+E]]

for line in tqdm(lines):
    if len(outputs) > 250_000:
        break

    packed = line.split("c9")
    fen = packed[0].strip()
    outcome = outcomescore[packed[1].strip().replace(";", "").replace("\"", "")]
    board = chess.Board(fen)
    
    phase = 0
    material = np.zeros(7)
    psqt = np.zeros((7, 64))

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

    kings = [-1, -1]



    for i in range(64):
        piece = board.piece_at(i)
        if piece == None:
            continue

        virtualboard[mailbox[i]] = (piece.piece_type * 2) + piece.color

        phase += phase_level[piece.piece_type]
        material[piece.piece_type] += sign[piece.color]

        j = i
        if piece.color:
            j = i ^ 56

        psqt[piece.piece_type][j] += sign[piece.color]

        if piece.piece_type == 6:
            kings[piece.color] = mailbox[i]

      



    #print(kings)
    #print(fen)
    #sys.exit()


    kingopen = np.zeros(1)

    for side in range(2):
        start = kings[side]
        for direction in patterns[6]:
            i = start + direction
            while virtualboard[i] == 0:
                i += direction
                kingopen += sign[side]
                  




    #attackers = np.zeros((14,14)) # TODO: add another table of attackers with tempo
    domesticmobilities = np.zeros(7)
    foreignmobilities = np.zeros(7)
    restrictedmobilities = np.zeros(7)

    for sq in range(len(virtualboard)):
        piece = virtualboard[sq]
        piecetype = piece // 2
        isray = rays[piecetype]
        pattern = patterns[piecetype]

        if piece == 2:
            #pattern = [S+W,S+E]
            pattern = []
        elif piece == 3:
            #pattern = [N+W,N+E]
            pattern = []
        
        domesticmobility = 0
        foreignmobility = 0
        restrictedmobility = 0

        for direction in pattern:
            current = sq
            for i in range(1,8):
                current += direction

                if virtualboard[current] != 0:
                    if virtualboard[current] == 1 or (piece & 1) == (virtualboard[current] & 1):
                        break
                


                domestic = (current >= 60)
                if piece & 1 == 1:
                    domestic = not domestic
                    if virtualboard[current + S + W] == 2 or virtualboard[current + S + E] == 2:
                        restrictedmobility += 1
                else:
                    if virtualboard[current + N + W] == 3 or virtualboard[current + N + E] == 3:
                        restrictedmobility += 1

                if domestic:
                    domesticmobility += 1
                else:
                    foreignmobility += 1

                if (not isray) or virtualboard[current] != 0:
                    break
        
        #if virtualboard[sq] > 1:
            #print("  pPnNbBrRqQkK"[virtualboard[sq]] , domesticmobility * sign[piece & 1], foreignmobility * sign[piece & 1])
        domesticmobilities[piecetype] += domesticmobility * sign[piece & 1]
        foreignmobilities[piecetype] += foreignmobility * sign[piece & 1]
        restrictedmobilities[piecetype] += restrictedmobility * sign[piece & 1]
        
    sidetomove = np.zeros(1)
    sidetomove[0] = sign[board.turn]


    
    if phase > 44:
        phase = 44
    
    start_weight = phase / 44
    end_weight = (44 - phase) / 44

    values = [material, psqt[1,:], domesticmobilities, foreignmobilities, restrictedmobilities, kingopen, sidetomove]
    if first:
        for item in values:
            sizes.append(item.shape[0])
        first = False

        print(fen)
        print(material)
        print("domestic",domesticmobilities)
        print("foreign",foreignmobilities)
        print("restricted",restrictedmobilities)
        print("kingvmob", kingopen)

    values = np.concatenate(values)    
    tapered_values = np.concatenate([values * start_weight, values * end_weight])

    inputs.append(tapered_values)
    outputs.append(outcome)


from sklearn.neural_network import MLPRegressor

mlp_regressor = MLPRegressor(hidden_layer_sizes=(1,), activation='logistic', verbose=True)

# Fit the MLPRegressor to your data
mlp_regressor.fit(inputs, outputs)

# Accessing the coefficients and intercept of the single-layer perceptron model
weights = mlp_regressor.coefs_[0]
bias = mlp_regressor.intercepts_[0]

half = len(weights)//2

def printcomma(idx):
    n1 = weights[idx][0]
    n2 = weights[idx + half][0]
    #print("{:.3f}".format(n), end=",")
    print("S("+str(int(128 * n1)) + "," + str(int(128*n2)) +")", end=",")

def printflat(start,end):
    for y in range(start,end):
        printcomma(y)
    print("")

def printpsqt(start,end,size=8): # m allows for adjusting psqt by material
    x = weights[start:end]
    i = 0
    for y in range(start,end):
        printcomma(y)
        if i % size == size-1:
            print("")
        i += 1

def plotpsqt(axs, start, label=""):
    x = (weights[start:start+64] * 128).reshape((8,8))
    y = (weights[start+half:start+half+64] * 128).reshape((8,8))

    axs[0].set_title(label)
    axs[0].imshow(x, vmin=-100, vmax=100)
    axs[1].imshow(y, vmin=-100, vmax=100)



print("Bias:", bias)
print("Weights:", weights)
print("Loss curve:", mlp_regressor.loss_curve_)
y_pred = mlp_regressor.predict(inputs)
residuals = outputs - y_pred
index_largest_residual = np.argmax(np.abs(residuals))
print("Largest residual:", outputs[index_largest_residual], y_pred[index_largest_residual], residuals[index_largest_residual], lines[index_largest_residual])
print("farout", np.sum(y_pred[y_pred < -0.1]) + np.sum(y_pred[y_pred > 1.1]))
print("r^2", mlp_regressor.score(inputs, outputs))

offset = 0
for size in sizes:
    #print(offset, size)
    if size == 64:
        fig, axs = plt.subplots(2,2)
        plotpsqt(axs[:,0],offset, "Pawns")
        plt.show()
    else:
        printflat(offset, offset+size)
    offset += size

