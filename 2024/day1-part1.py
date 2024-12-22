left_list = []
right_list = []

with open('day1-input.txt', 'r') as f:
    for line in f:
        left, right = line.split()
        left_list.append(int(left))
        right_list.append(int(right))


total_distance = 0
sorted_left = sorted(left_list)
sorted_right = sorted(right_list)

for i in range(len(left_list)):
    total_distance += abs(sorted_left[i] - sorted_right[i])

print(total_distance)
