import subprocess
import sys
from dataclasses import dataclass
from typing import List

@dataclass
class Tool:
    name: str
    install_url: str
    check_cmd: List[str]
    run_cmd: List[str]

TOOLS = [
    Tool(
        name="go-critic",
        install_url="github.com/go-critic/go-critic/cmd/gocritic@latest",
        check_cmd=["gocritic", "version"],
        run_cmd=[
            "gocritic",
            "check",
            "-enableAll",
            "-disable=commentedOutCode",
            "-@hugeParam.sizeThreshold=100",
        ],
    ),
    Tool(
        name="staticcheck",
        install_url="honnef.co/go/tools/cmd/staticcheck@latest",
        check_cmd=["staticcheck", "-version"],
        run_cmd=["staticcheck", "-checks", "all"],
    ),
]

def is_installed(tool: Tool) -> bool:
    try:
        subprocess.run(
            tool.check_cmd,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            check=True,
        )
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False

def install_tool(tool: Tool) -> None:
    print(f"Installing {tool.name}...")
    try:
        subprocess.run(
            ["go", "install", "-v", tool.install_url],
            check=True,
        )
    except subprocess.CalledProcessError:
        print(f"Failed to install {tool.name}.")
        sys.exit(1)

def main():
    print("Checking dependencies...")
    for tool in TOOLS:
        if not is_installed(tool):
            print(f"{tool.name} not found.")
            install_tool(tool)

    print("\nRunning tools...")
    exit_code = 0
    for tool in TOOLS:
        print(f"--- Running {tool.name} ---")
        result = subprocess.run(tool.run_cmd)
        if result.returncode != 0:
            print(f"{tool.name} found issues.")
            exit_code = 1
        print("\n")

    sys.exit(exit_code)

if __name__ == "__main__":
    main()
