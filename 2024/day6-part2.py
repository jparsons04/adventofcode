room = []
guard_directions = ['^', '>', 'v', '<']
direction_offsets = {'^': (-1, 0),
                     '>': (0, 1),
                     'v': (1, 0),
                     '<': (0, -1)}
visited = {}
num_obstructions_cause_loops = 0

with open('day6-input.txt', 'r') as f:
    for line in f:
        room_row = []
        for _, rr in enumerate(line.strip()):
            room_row.append(rr)
        room.append(room_row)


def move_guard(guardRow, guardCol, guardDir):
    room[guardRow][guardCol] = '.'
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
            if room[guardRow-1][guardCol] in ['#', 'O']:
                return '>'
        else:
            return 'out'
    elif guardDir == '>':
        # not on right edge of room
        if guardCol != len(room[0])-1:
            if room[guardRow][guardCol+1] in ['#', 'O']:
                return 'v'
        else:
            return 'out'
    elif guardDir == 'v':
        # not on bottom edge of room
        if guardRow != len(room)-1:
            if room[guardRow+1][guardCol] in ['#', 'O']:
                return '<'
        else:
            return 'out'
    elif guardDir == '<':
        # not on left edge of room
        if guardCol != 0:
            if room[guardRow][guardCol-1] in ['#', 'O']:
                return '^'
        else:
            return 'out'

    # no rotation needed, not on edge
    return guardDir


def get_next_location(guardRow, guardCol, guardDir):
    nextRow = guardRow + direction_offsets[guardDir][0]
    nextCol = guardCol + direction_offsets[guardDir][1]
    return room[nextRow][nextCol]


def rotate_until_safe(guardRow, guardCol, guardDir):
    directions = ['^', '>', 'v', '<']
    # Mark the visited direction before rotation
    visited.setdefault((guardRow, guardCol), set()).add(guardDir)

    while directions[0] != guardDir:
        _dir = directions.pop(0)
        directions.append(_dir)

    directions.remove(guardDir)
    guardDir = check_rotate(guardRow,
                            guardCol,
                            guardDir)
    if guardDir == 'out':
        return guardDir
    else:
        next_location = get_next_location(guardRow,
                                          guardCol,
                                          guardDir)
        # Verify that the next location is safe
        if next_location == '.':
            return guardDir
        else:
            # Rotate until it is safe
            while next_location != '.':
                directions.remove(guardDir)
                if len(directions) > 0:
                    guardDir = directions[0]
                    guardDir = check_rotate(guardRow,
                                            guardCol,
                                            guardDir)
                    next_location = get_next_location(guardRow,
                                                      guardCol,
                                                      guardDir)

            return guardDir


def find_guard():
    for row, rr in enumerate(room):
        for col, _ in enumerate(rr):
            if room[row][col] in guard_directions:
                return row, col, room[row][col]


if __name__ == "__main__":
    # Get guard's starting position and mark it as visited
    guardRow, guardCol, guardDir = find_guard()
    visited.setdefault((guardRow, guardCol), set()).add(guardDir)

    # Save the guard's original position
    # and direction so loop can be reset properly
    guardStartRow, guardStartCol, guardStartDir = guardRow, guardCol, guardDir

    loopDetected = False
    obstructionPlaced = False

    for row, rr in enumerate(room):
        for col, _ in enumerate(rr):
            # print('new loop:', row, col)
            # Can we place an obstruction here?
            # Cannot be where the guard starts
            # or where an obstruction alredy exists
            if room[row][col] not in ['#', '^']:
                room[row][col] = 'O'
                obstructionPlaced = True
            else:
                obstructionPlaced = False

            if obstructionPlaced is True:
                # print('obstruction is at:', row, col)
                # Move the guard with the obstruction in place
                # until there is a result
                while guardDir != 'out' and loopDetected is False:
                    guardDir = rotate_until_safe(guardRow,
                                                 guardCol,
                                                 guardDir)

                    if guardDir != 'out':
                        guardRow, guardCol = move_guard(guardRow,
                                                        guardCol,
                                                        guardDir)
                        visited_dirs = visited.get((guardRow, guardCol), set())
                        if guardDir in visited_dirs:
                            loopDetected = True
                            room[guardRow][guardCol] = '.'
                    else:
                        room[guardRow][guardCol] = '.'

                if loopDetected is True:
                    num_obstructions_cause_loops += 1

                # Remove the obstruction
                room[row][col] = '.'

                # Reset the loop
                visited = {}
                guardRow, guardCol, guardDir = \
                    guardStartRow, guardStartCol, guardStartDir
                loopDetected = False
                obstructionPlaced = False

    print(num_obstructions_cause_loops)
