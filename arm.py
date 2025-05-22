import requests
import json
import time
import numpy as np

def movearm(x,y,z=10,spd=0.25,t=3.14):
	requests.get("http://192.168.4.1/js?json=" + json.dumps({"T":104,"x":x,"y":y,"z":z,"t":t,"spd":spd}))

#      X+
#      
#  B-------D
#  |       |
#  |       |  Y-
#  |       |
#  A-------C
#     ARM
#      |
#     Y=0
A = (90,  150, -40)
B = (420, 160, -20)
C = (60, -165, -40)
D = (415,-160, -20)

def movesq(sq, heightOffset=0, spd=0.25,t=3.14, shift=0):
	dy = (ord(sq[0]) - ord('a'))/7
	dx = (ord(sq[1]) - ord('1'))/7
	dy = 1-(dy*2)
	print(dx, dy)

	if dy >= 0:
		x = (B[0]*dx) + (A[0]*(1-dx))
		y = (B[1]*dx) + (A[1]*(1-dx))
		z = (B[2]*dx) + (A[2]*(1-dx))
		y *= dy
	else:
		x = (D[0]*dx) + (C[0]*(1-dx))
		y = (D[1]*dx) + (C[1]*(1-dx))
		z = (D[2]*dx) + (C[2]*(1-dx))
		y *= abs(dy)

	if dx == 0:
		theta = np.arctan((0.5*dy)/0.00001) + (np.pi / 2)
	else:
		theta = np.arctan((0.5*dy)/dx) + (np.pi / 2)

	x += shift*np.cos(theta)
	y += shift*np.sin(theta)


	movearm(x, y, z=z+heightOffset, spd=spd, t=t)

def calibrate():
	movearm(A[0],A[1],A[2]) # A
	input()
	movearm(B[0],B[1], B[2]) #D
	input()
	movearm(C[0],C[1],C[2]) # C
	input()
	movearm(D[0],D[1],D[2]) # B


def calibrate2():
	movesq("a8")
	input()
	movesq("a8", shift=50)
	input()
	movesq("h1")
	input()
	movesq("h1", shift=50)
	input()
	movesq("e4")
	input()
	movesq("e4", shift=50)

def movemove(movestr):
	start = movestr[0:2]  
	end = movestr[2:4]
	movesq(start,t=2.7,spd=0.25)
	time.sleep(1)
	movesq(start,t=2.7,z=-65,spd=0.1)
	time.sleep(1)
	movesq(start,t=3.14,z=-65,spd=0.1)
	time.sleep(1)
	movesq(start,t=3.14)
	time.sleep(1)
	movesq(end,t=3.14)
	time.sleep(1)
	movesq(end,t=3.14,z=-60,spd=0.1)
	time.sleep(1)
	movesq(end,t=2.7,z=-60,spd=0.1)
	time.sleep(1)
	movesq(end,t=2.7)

#movemove("a1h8")

#calibrate()
calibrate2()



#movemove("a2a4")
#movemove("e2e4")
#movemove("h2h4")
#movemove("a4a2")
#movemove("e4e2")
#movemove("h4h2")


#movemove("h7h5")
#movemove("e7e5")
#movemove("a7a5")
#movemove("h5h7")
#movemove("e5e7")
#movemove("a5a7")

#movesq("_5")

