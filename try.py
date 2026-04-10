import matplotlib.pyplot as plt
import numpy as np

eg_file = np.repeat(np.array([-86, -37, -16, 0, -16, 2, -35, -74,]),8).reshape(8,8).transpose()
eg_rank = np.repeat(np.array([-20, 53, 78, 72, 74, 67, 47, 0,][::-1]),8).reshape(8,8)

eg_table = eg_file + eg_rank


mg_file = np.repeat(np.array([152, 32, 58, 0, 133, 51, 162, 253]),8).reshape(8,8).transpose()
mg_rank = np.repeat(np.array([199, 87, 53, -83, -160, -167, -94, 0, ][::-1]),8).reshape(8,8)

mg_table = mg_file + mg_rank

table = np.array([
-119, -461,  -60,  158,  429,  412, 1397, 1163,
 204,  370,  611,  272,  442,  300,  680,  602,
 459,  544,  484,  272,  477,  671,  522,  380,
 408,  570,  250,  495,  477,  403,  327,  384,
 291,  236,  300,  505,  392,  343,  216,  163,
 -62,  194,  167,  184,  252,  289,  422,   64,
   4,  200,  398,  265,  188,  451,  330,   86,
 162,  213,  321,   80,  275,  593,   71,  131,
]).reshape(8,8)

mobtable = np.zeros((8,8))

for y in range(8):
	for x in range(8):
		mobtable[x][y] += table[x][y]
		if x-1 >= 0:
			mobtable[x][y] += table[x-1][y]
			if y-1 >= 0:
				mobtable[x][y] += table[x-1][y-1]
			if y+1 <= 7:
				mobtable[x][y] += table[x-1][y+1]
		if x+1 <= 7:
			mobtable[x][y] += table[x+1][y]
			if y-1 >= 0:
				mobtable[x][y] += table[x+1][y-1]
			if y+1 <= 7:
				mobtable[x][y] += table[x+1][y+1]
		
		if y-1 >= 0:
			mobtable[x][y] += table[x][y-1]
		if y+1 <= 7:
			mobtable[x][y] += table[x][y+1]


plt.imshow(mg_table)
plt.show()