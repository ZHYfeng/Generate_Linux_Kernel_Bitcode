import json
import subprocess
from concurrent.futures import ThreadPoolExecutor
from tqdm import tqdm


def preprocess_command(command):
    modified_command = command
    # Modify the command as needed
    flag = ''
    flag += ' -mllvm -disable-llvm-optzns'
    flag += ' -fno-discard-value-names'
    flag += ' -emit-llvm'
    flag += ' -w -g -c '
    modified_command = modified_command.replace(' -c ', flag)
    modified_command = modified_command.replace('.o ', '.bc ')
    modified_command = modified_command.replace(' -Os ', ' -O0 ')
    modified_command = modified_command.replace(' -O3 ', ' -O0 ')
    modified_command = modified_command.replace(' -O2 ', ' -O0 ')
    modified_command = modified_command.replace(
        ' -fno-var-tracking-assignments ', ' ')
    modified_command = modified_command.replace(' -fconserve-stack ', ' ')
    modified_command = modified_command.replace(' -march=armv8-a+crypto ', ' ')
    modified_command = modified_command.replace(' -mno-fp-ret-in-387 ', ' ')
    modified_command = modified_command.replace(' -mskip-rax-setup ', ' ')
    modified_command = modified_command.replace(
        ' -ftrivial-auto-var-init=zero ', ' ')
    return modified_command


def execute_command(command):
    if not command.startswith('clang '):
        return
    modified_command = preprocess_command(command)
    # print(modified_command)
    subprocess.run(modified_command, shell=True)


def execute_commands_concurrently(commands, max_workers):
    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        results = list(tqdm(executor.map(execute_command, commands),
                       total=len(commands), desc='Executing Commands'))


# Read compile_commands.json file
with open('compile_commands.json', 'r') as file:
    compile_commands = json.load(file)

# Extract commands from compile_commands.json
commands = [entry['command'] for entry in compile_commands]

# Adjust the number of concurrent threads
max_workers = 32  # Example: Set the maximum number of concurrent threads to 4

# Execute the commands concurrently with a progress bar
execute_commands_concurrently(commands, max_workers)
