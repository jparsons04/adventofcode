import re

pattern = re.compile(r'mul\((\d{1,3})\,(\d{1,3})\)')
total = 0

with open('day3-input.txt', 'r') as f:
    for line in f:
        for m in re.finditer(pattern, line):
            total += (int(m.group(1)) * int(m.group(2)))

print(total)
