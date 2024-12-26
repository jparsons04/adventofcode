room = []
guard_directions = ['^', '>', 'v', '<']

with open('day6-input.txt', 'r') as f:
    for line in f:
        room_row = []
        for _, rr in enumerate(line.strip()):
            room_row.append(rr)
        room.append(room_row)


def move_guard(guardRow, guardCol, guardDir):
    if guardDir == '^':
        room[guardRow-1][guardCol] = '^'
        return guardRow-1, guardCol
    elif guardDir == '>':
        room[guardRow][guardCol+1] = '>'
        return guardRow, guardCol+1
    elif guardDir == 'v':
        room[guardRow+1][guardCol] = 'v'
        return guardRow+1, guardCol
    elif guardDir == '<':
        room[guardRow][guardCol-1] = '<'
        return guardRow, guardCol-1


def check_rotate(guardRow, guardCol, guardDir):
    if guardDir == '^':
        # not on top edge of room
        if guardRow != 0:
            if room[guardRow-1][guardCol] == '#':
                return '>'
        else:
            return 'out'
    elif guardDir == '>':
        # not on right edge of room
        if guardCol != len(room[0])-1:
            if room[guardRow][guardCol+1] == '#':
                return 'v'
        else:
            return 'out'
    elif guardDir == 'v':
        # not on bottom edge of room
        if guardRow != len(room)-1:
            if room[guardRow+1][guardCol] == '#':
                return '<'
        else:
            return 'out'
    elif guardDir == '<':
        # not on left edge of room
        if guardCol != 0:
            if room[guardRow][guardCol-1] == '#':
                return '^'
        else:
            return 'out'
    return guardDir


def find_guard():
    for row, rr in enumerate(room):
        for col, _ in enumerate(rr):
            if room[row][col] in guard_directions:
                return row, col, room[row][col]


if __name__ == "__main__":
    # Get guard's starting position and mark it as visited
    guardRow, guardCol, guardDir = find_guard()

    while guardDir != 'out':
        guardDir = check_rotate(guardRow, guardCol, guardDir)

        room[guardRow][guardCol] = 'X'

        if guardDir != 'out':
            guardRow, guardCol = move_guard(guardRow, guardCol, guardDir)

    print(sum(x.count('X') for x in room))
