reports = []


def level_safety_check(lvl1, lvl2, increasing):
    if (lvl2 > lvl1 and increasing is True) \
            or (lvl2 < lvl1 and increasing is False):
        if abs(lvl2 - lvl1) >= 1 and abs(lvl2 - lvl1) <= 3:
            return True
        else:
            return False
    else:
        return False


def process_report(report, options):
    cur_level = None
    previous_level = None
    increasing = None

    safe = options['safe']
    dampened = options['dampened']

    local_report = report
    print(f'Local report: {local_report}')

    for i in range(len(local_report)):
        cur_level = local_report[i]
        if i == 0:
            pass
        elif i == 1:
            increasing = cur_level > previous_level
            if level_safety_check(previous_level, cur_level, increasing) is False:
                if dampened is False:
                    if i != len(local_report)-1:
                        unsafe_right = local_report[i+1]
                        if increasing is True:
                            if local_report[i] < unsafe_right:
                                print(f'Inc, Level {local_report[i-1]} is unsafe, dampening')
                                del local_report[i-1]
                            else:

                                print(f'Inc, Level {local_report[i]} is unsafe, dampening')
                                del local_report[i]
                        else:
                            if local_report[i] > unsafe_right:
                                print(f'Dec, Level {local_report[i-1]} is unsafe, dampening')
                                del local_report[i-1]
                            else:
                                print(f'Dec, Level {local_report[i]} is unsafe, dampening')
                                del local_report[i]
                    else:
                        print(f'End of list, Level {local_report[i]} is unsafe, dampening')
                        del local_report[i]
                    options['safe'] = False
                    options['do_dampen'] = True
                    return local_report, options
                else:
                    print(f'Level {local_report[i]} is unsafe, already dampened\n')
                    options['safe'] = False
                    return local_report, options
        else:
            cur_level = local_report[i]
            check_result = level_safety_check(previous_level,
                                              cur_level,
                                              increasing)
            if check_result is False:
                if dampened is False:
                    if i != len(local_report)-1:
                        # Monotonically increasing
                        if local_report[i-2] < local_report[i-1] and local_report[i-1] < local_report[i]:
                            if abs(local_report[i-2] - local_report[i-1]) > 3:
                                del local_report[i-1]
                            else:
                                del local_report[i]
                        # Not monotonically increasing
                        else:
                            # Monotonically decreasing
                            if local_report[i-2] > local_report[i-1] and local_report[i-1] > local_report[i]:
                                if abs(local_report[i-2] - local_report[i-1]) > 3:
                                    del local_report[i-1]
                                else:
                                    del local_report[i]
                            # Neither monotonically increasing nor decreasing
                            else:
                                # Report increasing
                                if local_report[i] > local_report[i-2] and local_report[i+1] > local_report[i]:
                                    del local_report[i-1]
                                # Report increasing
                                elif local_report[i-1] > local_report[i-2] and local_report[i+1] > local_report[i-1]:
                                    del local_report[i]
                                # Report decreasing
                                elif local_report[i] < local_report[i-2] and local_report[i+1] < local_report[i]:
                                    del local_report[i-1]
                                # Report decreasing
                                elif local_report[i-1] < local_report[i-2] and local_report[i+1] < local_report[i-1]:
                                    del local_report[i]
                    else:
                        print(f'End of list, Level {local_report[i]} is unsafe, dampening')
                        del local_report[i]

                    options['safe'] = False
                    options['do_dampen'] = True
                    return local_report, options
                else:
                    print(f'Level {local_report[i]} is unsafe, already dampened\n')
                    options['safe'] = False
                    return local_report, options
        previous_level = cur_level

    options['safe'] = True
    print(f'Report {local_report} is SAFE, {safe}')
    return local_report, options


with open('day2-input.txt', 'r') as f:
    for line in f:
        reports.append(line.strip())


safe_reports = 0

for i in range(len(reports)):
    report = [int(x) for x in reports[i].split()]

    options = {'safe': True,
               'dampened': False,
               'do_dampen': False}

    print(f'Report {i+1}...')
    report, options, = process_report(report, options)
    if options['safe'] is True and options['do_dampen'] is False:
        safe_reports += 1
        print(f'Safe reports: {safe_reports}\n')
    # Re-run level analysis after unsafe level is removed
    elif options['safe'] is False and options['do_dampen'] is True:
        options['dampened'] = True
        options['safe'] = True
        report, options = process_report(report, options)
        if options['safe'] is True:
            safe_reports += 1
            print(f'Safe reports: {safe_reports}\n')

print(safe_reports)
