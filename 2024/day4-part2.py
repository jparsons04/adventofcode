word_search = []

adjacent = [[-1, -1], [-1, 1],
            [1, -1], [1, 1]]

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


def find_x_mas(y, x):
    maxlen = len(word_search[0])-1
    found_m = []
    found_s = []
    if x > 0 and x < maxlen and y > 0 and y < maxlen:
        found_m = lookup_surrounding(range(4), 'M')
        found_s = lookup_surrounding(range(4), 'S')

        # The M-A-S in a cross means there are exactly two M's and two S's
        if len(found_m) == 2 and len(found_s) == 2:
            # Want to avoid a diagonal of M-A-M
            if (found_m[0][0] == 0 and found_m[1][0] == 3) \
                    or (found_m[0][0] == 1 and found_m[1][0] == 2):
                return False
            else:
                return True
        else:
            return False
    else:
        return False


for y, line in enumerate(word_search):
    for x, char in enumerate(line):
        if char == 'A':
            if find_x_mas(y, x) is True:
                xmas_matches += 1

print(xmas_matches)
