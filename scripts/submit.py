#!/usr/bin/env python

import argparse
import datetime

import requests


def parse():
    types = ['limit', 'market']
    sides = ['buy', 'sell']

    parser = argparse.ArgumentParser()

    # parser.add_argument('-s', '--symbol')
    parser.add_argument('-d', '--side', default='buy', choices=sides)
    parser.add_argument('-i', '--id', default=datetime.datetime.now().isoformat())
    parser.add_argument('-p', '--price', default='0')
    parser.add_argument('-q', '--quantity', default='1')
    parser.add_argument('-t', '--type', default='limit', choices=types)

    args = parser.parse_args()
    args.type = types.index(args.type)
    args.side = sides.index(args.side)
    return args


def main():
    args = parse()
    resp = requests.post('http://127.0.0.1:7701/orders/', json=args.__dict__)
    print('Status code:')
    print(' ', resp.status_code)
    print('Response:')
    print(' ', resp.json())


if __name__ == '__main__':
    main()
