#!/usr/bin/env python

import sys
import argparse

version = '%(prog)s 1.0'

def ArgParse():
    parser = argparse.ArgumentParser(description=
    '''This script compares lists.
    
    1: Values only include in File 1
    2: Values include in File 1 and File 2
    3: Values only include in File 2

    Create Date: 2016-11-15 ''',
    formatter_class=argparse.RawDescriptionHelpFormatter)
    
    parser.add_argument('infile1', nargs='?', type=argparse.FileType('rU'), metavar='FILE1', help='List File 1', default=sys.stdin)
    parser.add_argument('infile2', nargs='?', type=argparse.FileType('rU'), metavar='FILE2', help='List File 2', default=sys.stdin)
    parser.add_argument('-v', '--version', action='version', version=version)
    args = parser.parse_args()

    return args

if __name__ == "__main__":
    args = ArgParse()
    cmplist = {}
    for line in args.infile1:
        key = line.rstrip()
        cmplist[key] = 1

    for line in args.infile2:
        key = line.rstrip()
        if cmplist.has_key(key):
            cmplist[key] = 2
        else:
            cmplist[key] = 3

    for key, value in cmplist.items():
        print("%s,%s" % (key, value))
