#!/usr/bin/env python

import sys
import yaml
import os
from datetime import datetime

def change_to_markdown(change):
    return '\n'.join("- {}".format(line.strip()) for line in change.strip().split('\n'))

def format_date(date_str):
    date_formats = ["%b %d, %Y", "%B %d, %Y"]
    for date_format in date_formats:
        try:
            return datetime.strptime(date_str, date_format).date()
        except ValueError:
            pass  # If the format doesn't match, move to the next one
    
    raise ValueError(f"Date string '{date_str}' does not match any supported format.")

def capitalize(name):
    fixed_mapping = {
        'ir': 'IR',
        'api': 'API',
        'xds': 'xDS',
        'ci-tooling-testing': 'CI Tooling Testing',
    }
    if name in fixed_mapping:
        return fixed_mapping[name]
    return name.capitalize()

def convert_yaml_to_markdown(input_yaml_file, output_markdown_path):
    # Extract the title from the input file name
    title = os.path.basename(input_yaml_file).split('.yaml')[0]
    # Generate the filename of output markdown file.
    output_markdown_file=os.path.join(output_markdown_path,title + '.md')

    with open(input_yaml_file, 'r') as file:
        data = yaml.safe_load(file)

    with open(output_markdown_file, 'w') as file:
        file.write('---\n')
        file.write('title: "{}"\n'.format(title))
        file.write("publishdate: {}\n".format(format_date(data['date'])))
        file.write('---\n\n')

        file.write("Date: {}\n\n".format(data['date']))

        for area in data['changes']:
            file.write("## {}\n".format(capitalize(area['area'])))
            if 'change' in area:
                file.write(change_to_markdown(area['change']) + '\n\n')

            if 'breaking-change' in area:
                file.write("### Breaking Changes\n")
                file.write(change_to_markdown(area['breaking-change']) + '\n\n')

    print("Markdown file '{}' has been generated.".format(output_markdown_file))

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python yml2md.py <input_yaml_file> <output_markdown_path>")
        sys.exit(1)

    input_yaml_file = sys.argv[1]
    output_markdown_path = sys.argv[2]
    convert_yaml_to_markdown(input_yaml_file, output_markdown_path)
