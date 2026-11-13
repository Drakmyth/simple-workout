"""Generate the short, gently enveloped two-tone rest chime (standard library only)."""

import math
from pathlib import Path
import struct
import wave

RATE = 16000
DURATION = 0.7
OUTPUT = Path(__file__).resolve().parents[2] / "assets/static/rest-chime.wav"

samples = bytearray()
for index in range(int(RATE * DURATION)):
    time = index / RATE
    value = 0.0
    for start, frequency in ((0.0, 660), (0.22, 880)):
        elapsed = time - start
        if 0 <= elapsed < 0.4:
            envelope = min(elapsed / 0.015, 1) * (1 - elapsed / 0.4) ** 2
            value += 0.18 * envelope * math.sin(2 * math.pi * frequency * elapsed)
    samples.extend(struct.pack("<h", round(value * 32767)))

with wave.open(str(OUTPUT), "wb") as output:
    output.setnchannels(1)
    output.setsampwidth(2)
    output.setframerate(RATE)
    output.writeframes(samples)
