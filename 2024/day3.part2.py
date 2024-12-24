import re

mul_pattern = re.compile(r'mul\((\d{1,3})\,(\d{1,3})\)')
do_pattern = re.compile(r'do\(\)')
dont_pattern = re.compile(r'don\'t\(\)')
total = 0

with open('day3-input.txt', 'r') as f:
    mul_on = True

    for line in f:
        dos = []
        donts = []

        keep_on = None

        for m in re.finditer(do_pattern, line):
            dos.append(m.start())

        for m in re.finditer(dont_pattern, line):
            donts.append(m.start())

        if len(dos) > 0:
            do_pos = dos.pop(0)

        if len(donts) > 0:
            dont_pos = donts.pop(0)

        for m in re.finditer(mul_pattern, line):
            next_do = None
            next_dont = None

            # dont_pos < m.start() < do_pos
            if m.start() > dont_pos and m.start() < do_pos:
                # print(f'do: {do_pos}, dont: {dont_pos}')
                if keep_on is not True:
                    # print('switching off')
                    mul_on = False
                while m.start() > dont_pos and len(donts) > 0:
                    dont_pos = donts.pop(0)
                # print(f'after... do: {do_pos}, dont: {dont_pos}')
            # do_pos < m.start() < dont_pos
            elif m.start() > do_pos and m.start() < dont_pos:
                # print(f'do: {do_pos}, dont: {dont_pos}')
                if keep_on is not False:
                    # print('switching on')
                    mul_on = True
                while m.start() > do_pos and len(dos) > 0:
                    do_pos = dos.pop(0)
                # print(f'after... do: {do_pos}, dont: {dont_pos}')
            # dd_pos < dont_pos < m.start()
            elif do_pos < dont_pos and dont_pos < m.start():
                # print('m.start() > both do_pos and dont_pos, switching off')
                # print(f'do: {do_pos}, dont: {dont_pos}')
                mul_on = False
                if len(dos) == 0:
                    # print('keep_on False')
                    keep_on = False
                while m.start() > dont_pos and len(donts) > 0:
                    dont_pos = donts.pop(0)
                while m.start() > do_pos and len(dos) > 0:
                    do_pos = dos.pop(0)
                # print(f'after... do: {do_pos}, dont: {dont_pos}')
            elif dont_pos < do_pos and do_pos < m.start():
                # print('m.start() > both do_pos and dont_pos, switching on')
                # print(f'do: {do_pos}, dont: {dont_pos}')
                mul_on = True
                if len(donts) == 0:
                    # print('keep_on True')
                    keep_on = True
                while m.start() > dont_pos and len(donts) > 0:
                    dont_pos = donts.pop(0)
                while m.start() > do_pos and len(dos) > 0:
                    do_pos = dos.pop(0)
                # print(f'after... do: {do_pos}, dont: {dont_pos}')

            if mul_on is True:
                total += (int(m.group(1)) * int(m.group(2)))
            # print(f'total now is {total}')

print(total)
