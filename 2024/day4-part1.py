word_search = []

adjacent = [[-1, -1], [-1, 0], [-1, 1],
            [0, -1], [0, 1],
            [1, -1], [1, 0], [1, 1]]

xmas_matches = 0

with open('day4-input.txt', 'r') as f:
    for _, line in enumerate(f):
        word_search_line = []
        for _, c in enumerate(line.strip()):
            word_search_line.append(c)
        word_search.append(word_search_line)


def lookup_surrounding(points, letter):
    found = []
    for i in points:
        if word_search[y + adjacent[i][0]][x + adjacent[i][1]] == letter:
            found.append((i, [y + adjacent[i][0], x + adjacent[i][1]]))
    return found


def count_potential_matches(m_list):
    xmas_count = 0

    for direction, position in m_list:
        direction_offset = adjacent[direction]
        a_pos = [position[0] + direction_offset[0],
                 position[1] + direction_offset[1]]
        if a_pos[0] >= 0 and a_pos[0] < len(word_search[0]) \
                and a_pos[1] >= 0 and a_pos[1] < len(word_search[0]):
            if word_search[a_pos[0]][a_pos[1]] == 'A':
                # print(f'A found at {a_pos[0]}, {a_pos[1]}')
                s_pos = [a_pos[0] + direction_offset[0],
                         a_pos[1] + direction_offset[1]]
                if s_pos[0] >= 0 and s_pos[0] < len(word_search[0]) \
                        and s_pos[1] >= 0 and s_pos[1] < len(word_search[0]):
                    if word_search[s_pos[0]][s_pos[1]] == 'S':
                        # print(f'S found at {s_pos[0]}, {s_pos[1]}')
                        xmas_count += 1
    # print(f'found {xmas_count} matches')
    return xmas_count


def find_m(y, x):
    found_m = []
    maxlen = len(word_search[0])-1
    if x > 0 and x < maxlen:
        # can do full adjacent search
        if y > 0 and y < maxlen:
            found_m = lookup_surrounding(range(8), 'M')
        # top edge (not corner)
        elif y == 0:
            found_m = lookup_surrounding(range(3, 8), 'M')
        # bottom edge (not corner)
        elif y == maxlen:
            found_m = lookup_surrounding(range(0, 5), 'M')
    # left edge (not corner)
    elif x == 0 and y > 0 and y < maxlen:
        found_m = lookup_surrounding([1, 2, 4, 6, 7], 'M')
    # right edge (not corner)
    elif x == maxlen and y > 0 and y < maxlen:
        found_m = lookup_surrounding([0, 1, 3, 5, 6], 'M')
    # upper left corner
    elif x == 0 and y == 0:
        found_m = lookup_surrounding([4, 6, 7], 'M')
    # upper right corner
    elif x == maxlen and y == 0:
        found_m = lookup_surrounding([3, 5, 6], 'M')
    # lower left corner
    elif x == 0 and y == maxlen:
        found_m = lookup_surrounding([1, 2, 4], 'M')
    # lower right corner
    elif x == maxlen and y == maxlen:
        found_m = lookup_surrounding([0, 1, 3], 'M')

    return found_m


for y, line in enumerate(word_search):
    for x, char in enumerate(line):
        if char == 'X':
            # print('X:', y, x)
            m_list = find_m(y, x)
            # print(m_list)
            xmas_matches += count_potential_matches(m_list)

print(xmas_matches)
