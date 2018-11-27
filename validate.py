
import math
import json
import sys
import subprocess
import sys
import os
import time
import argparse

ON_POSIX = 'posix' in sys.builtin_module_names


ref = sys.argv[1]

json_in = open(sys.argv[2], "r")
a = json.load(json_in)

suc = a['success']

print("Validating: ", sys.argv[2])

frames = suc['gameResult']['frames']
for x in frames:
    x['view'] = x['view'].split("\n")

setup = frames[0]['view']
# strip the extra startup data
frames[0]['view'] = [frames[0]['view'][0]] + frames[0]['view'][5:]


referee = subprocess.Popen([ref], stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                           bufsize=1, universal_newlines=True, close_fds=ON_POSIX)


# startup command
print("###Validate", file=referee.stdin)

# create checkpoint list
cpsstr = [int(x) for x in setup[3].split(" ")]
cps = []
for x in range(0, len(cpsstr), 2):
    cps.append((cpsstr[x], cpsstr[x+1]))
print(len(cps), file=referee.stdin)
for cp in cps:
    print(cp[0], cp[1], file=referee.stdin)


def single_pod_input(pod_line):
    p = pod_line.split(" ")
    x = int(float(p[0]))
    y = int(float(p[1]))
    vx = int(float(p[2]))
    vy = int(float(p[3]))
    if p[8] == "null":
        ang = -1
    else:
        ang = round(math.degrees(float(p[8])))
        while ang < 0:
            ang += 360
        while ang > 360:
            ang -= 360
    ncp = int(p[10])
    return [x, y, vx, vy, ang, ncp]


def get_ref_input(data):
    return [single_pod_input(data[1]), single_pod_input(data[3]), single_pod_input(data[5]), single_pod_input(data[7])]


cur_input = (get_ref_input(frames[0]['view']))


for x in range(2):
    ignore = referee.stdout.readline()
    ignore = referee.stdout.readline()
    ref_ncp = int(referee.stdout.readline())
    for v in range(ref_ncp):
        cp = [int(x) for x in referee.stdout.readline().split()]
        if (cp[0] != cps[v][0]) or (cp[1] != cps[v][1]):
            print(f"CHECKPOINT ERROR {v} ({cps[v]}) != ({cp})")
            exit(-1)

first = True
frames = frames[1:]

turn = 0
results = [0, 0]
lastinput = cur_input
curoutput = ["", "", "", ""]
lastoutput = ["", "", "", ""]
outputind = 0
while True:

    # ignore ###Input
    ignore = referee.stdout.readline().split()
    if ignore[0] == "###End":
        results = [int(x) for x in ignore[1:]]
        break
    fr = frames[turn]
    inlines = []

    for x in range(4):
        inlines.append(referee.stdout.readline().strip())
    error = False
    for input_pos, input_val in enumerate(cur_input):
        if turn % 2 > 0:
            # order for p2 is 2, 3, 0, 1
            input_pos = (input_pos + 2) % 4
        v = [int(x) for x in inlines[input_pos].split()]

        for pos, val in enumerate(v):
            if turn > 1 and pos == 4:
                while val < 0:
                    val += 360
                while val >= 360:
                    val -= 360
            if input_val[pos] != val:
                print(
                    f"Input error turn {turn} rec:'{val}' exp:'{input_val[pos]}' index: {pos} pod: {input_pos}")
                error = True
    if error:
        print(len(cps), file=sys.stderr)
        for x in cps:
            print(x[0], x[1], file=sys.stderr)
        for x in lastinput:
            print(" ".join(map(str, x)), file=sys.stderr)
        for x in lastoutput:
            print(x, file=sys.stderr)
        exit(-1)
    # ignore ###Output
    ignore = referee.stdout.readline().split()
    if ignore[0] == "###End":
        results = [int(x) for x in ignore[1:]]
        break
    if 'stdout' not in fr:
        if turn % 2 > 0:
            results = [0, 1]
        else:
            results = [1, 0]
        break
    output = fr['stdout'].split("\n")
    if len(output) < 3:
        if turn % 2 > 0:
            results = [0, 1]
        else:
            results = [1, 0]
        break

    for outline in output[:2]:
        curoutput[outputind] = outline
        outputind += 1
        print(outline.strip(), file=referee.stdin)

    if fr['keyframe']:
        outputind = 0
        lastinput = cur_input
        lastoutput = curoutput
        cur_input = get_ref_input(fr['view'])
    turn += 1
actual_results = suc['gameResult']['ranks']
for p, x in enumerate(results):
    if actual_results[p] != x:
        print(f"RESULT ERROR: exp: {actual_results}  rec: {results}")
        exit(0)

print(sys.argv[2], " is valid")

