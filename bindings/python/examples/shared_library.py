# Derived from tui-test (https://github.com/microsoft/tui-test), MIT, Microsoft Corporation.
"""Run with an installed wheel or NASDUCK_NATIVE_LIBRARY."""

import asyncio
import sys

from nasduck import Nasduck


async def main():
    async with Nasduck.ephemeral("shared-library") as terminal:
        await terminal.run(
            sys.executable, "-u", "-c",
            "print('Shared typed ABI: Python'); input()",
        )
        await terminal.get_by_text("Shared typed ABI: Python").expect()
        packed, cols, rows = await terminal._packed_screen()
    assert packed.readonly
    assert b"Shared typed ABI: Python" in packed.tobytes()
    print("Shared library: {} columns, {} rows".format(cols, rows))
    print("Copied screen remains valid after close:")
    print(packed.tobytes().decode("utf-8").strip())


if __name__ == "__main__":
    asyncio.run(main())
