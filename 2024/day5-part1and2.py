import functools

ordering_rules = {}
updates = []

correctly_ordered_page_total = 0
incorrectly_ordered_page_total = 0

with open('day5-input.txt', 'r') as f:
    parsing_rules = True
    for line in f:
        if "|" in line:
            rule = line.strip().split('|')
            ordering_rules.setdefault(rule[0], []).append(rule[1])
        elif "," in line:
            updates.append(line.strip().split(','))


def sort_pages(page1, page2):
    if ordering_rules.get(page1, -1) == -1:
        ordering_rules[page1] = []

    if ordering_rules.get(page2, -1) == -1:
        ordering_rules[page2] = []

    if page2 in ordering_rules[page1]:
        return -1
    elif page1 in ordering_rules[page2]:
        return 1


for pages in updates:
    ordered_pages = sorted(pages, key=functools.cmp_to_key(sort_pages))
    middle_page = int(ordered_pages[len(ordered_pages)//2])

    # For day 5, part 1
    if ordered_pages == pages:
        # print('correctly sorted updates:', pages)
        correctly_ordered_page_total += middle_page
    # For day 5, part 2
    else:
        # print('resorted updates:', ordered_pages)
        incorrectly_ordered_page_total += middle_page

print(correctly_ordered_page_total)
print(incorrectly_ordered_page_total)
