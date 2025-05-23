#!/usr/bin/env python3
"""
Builds schema definitions in ../ from all templates in this directory.
It will resolve all references and remove the $defs section before writing
the final json schema.
Existing files get overwritten!

Usage: FORCE=1 ./build.py

"""

from copy import deepcopy
import glob
import os
import sys
import jinja2
import json
import jsonref
import pprint


def main():

    # Check if FORCE is set.
    try:
        if os.environ['FORCE'] != '1':
            raise ValueError
    except (KeyError, ValueError):
        print('Variable FORCE not set to 1, so we terminate.', file=sys.stderr)
        sys.exit(1)

    # Set up Jinja2 environment. 
    templateLoader = jinja2.FileSystemLoader(searchpath='.')
    templateEnv = jinja2.Environment(loader=templateLoader)

    # Walk through all templates, generate the schema filr and write it into the parent directory.
    fail = False
    for file in glob.glob('saptune*.template'):
        src = file
        dest = f'../{file}'.rstrip('.template')

        try:
            template = templateEnv.get_template(file)
            schema_dict = jsonref.loads(template.render())    # Render the template, load the JSON
            schema_dict_touched = deepcopy(schema_dict)   # We have to touch the properties to resolve references.
            del schema_dict_touched["$defs"]  # With references resolved, we don't need '$defs' anymore. 
            schema = json.dumps(schema_dict_touched, indent=4)   # Format JSON nicely and write out.
            with open(dest, 'w') as f:
                f.write(schema)
                print(f'[\033[32m OK \033[39m] "{src}" -> "{dest}"')
        except Exception as err:
            print(f'[\033[31mFAIL\033[39m] "{file}": {err}', file=sys.stderr)
            fail = True

    # Bye.
    sys.exit(2) if fail else sys.exit(0)


if __name__ == '__main__':
    main()

