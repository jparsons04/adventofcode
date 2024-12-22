left_list = []
right_list = []

with open('day1-input.txt', 'r') as f:
    for line in f:
        left, right = line.split()
        left_list.append(int(left))
        right_list.append(int(right))


total_similarity = 0

for i in range(len(left_list)):
    total_similarity += left_list[i] * right_list.count(left_list[i])

print(total_similarity)
