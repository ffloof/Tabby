import requests
import json
import time

print()

x1 = 110
y1 = 160
z1 = -70

size_x = 250
size_y = -250
size_z = 0





#290
#220

def movearm(x,y,z=10,spd=0.25,t=3.14):
	requests.get("http://192.168.4.1/js?json=" + json.dumps({"T":104,"x":x,"y":y,"z":z,"t":t,"spd":spd}))


def movesq(sq, z=10,spd=0.25,t=3.14 ,shiftx=0, shifty=0):
	dy = (ord(sq[0]) - ord('a'))
	dx = (ord(sq[1]) - ord('1'))

	sqx = x1 + (size_x * (dx / 7)) - (size_x / 16) + shiftx
	sqy = y1 + (size_y * (dy / 7)) - (size_y / 16) + shifty
	movearm(sqx, sqy, z=z, spd=spd, t=t)

	mag = ((sqx ** 2) + (sqy ** 2)) ** 0.5
	print("deflection", sqx/mag, sqy/mag)



def calibrate():
	movearm(x1,y1,z=z1)
	input()
	#movearm(x1+size_x,y1,z=z1+size_z)
	#time.sleep(3)
	movearm(x1+size_x,y1+size_y,z=z1+size_z)
	input()
	movearm(x1,y1+size_y,z=z1+size_z)
	input()


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

movesq("a1", z=100)
calibrate()

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

