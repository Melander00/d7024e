#!/usr/bin/env python3

"""Analyze lookup reliability as a function of simulated packet loss.

Usage:
    python analyze_packet_loss.py data/packet_loss.txt
    python analyze_packet_loss.py data/packet_loss.txt --output data/packet_loss_results
"""

import argparse
import csv
import re
import statistics
from collections import defaultdict
from pathlib import Path

import matplotlib.pyplot as plt


EXPERIMENT_RE = re.compile(
    r"^#\s*Experiment\s+nr:\s*(\d+)"
    r"\s+nodes=(\d+)"
    r"\s+packet_loss=([0-9.]+)"
    r"\s+seed=(\d+)"
    r"\s+lookups=(\d+)"
)
RESULT_RE = re.compile(
    r"^\S+\s+lookup_result\s+(\d+)\s+success=(true|false)"
)


def parse_log(filename):
    measurements = []
    # The log contains multiple headers, so parse each block independently.
    current = None
    with open(filename, "r", encoding="utf-8") as stream:
        for line_number, raw_line in enumerate(stream, start=1):
            line = raw_line.strip()
            header = EXPERIMENT_RE.match(line)
            if header:
                current = {
                    "experiment": int(header.group(1)),
                    "nodes": int(header.group(2)),
                    "packet_loss": float(header.group(3)),
                    "seed": int(header.group(4)),
                    "expected_lookups": int(header.group(5)),
                    "results": [],
                }
                measurements.append(current)
                continue

            result = RESULT_RE.match(line)
            if result and current is not None:
                current["results"].append(
                    {
                        "lookup": int(result.group(1)),
                        "success": result.group(2) == "true",
                    }
                )
            elif result:
                print(f"WARNING: result before header at line {line_number}")

    return measurements


def summarize(measurements):
    grouped = defaultdict(list)
    rows = []

    for measurement in measurements:
        results = measurement["results"]
        expected = measurement["expected_lookups"]
        if len(results) != expected:
            print(
                "WARNING: packet_loss=%.2f seed=%d expected %d results, got %d"
                % (
                    measurement["packet_loss"],
                    measurement["seed"],
                    expected,
                    len(results),
                )
            )

        successes = sum(result["success"] for result in results)
        attempts = len(results)
        rate = successes / attempts if attempts else 0.0
        grouped[measurement["packet_loss"]].append(rate)
        rows.append(
            {
                "packet_loss": measurement["packet_loss"],
                "seed": measurement["seed"],
                "nodes": measurement["nodes"],
                "lookups": attempts,
                "successes": successes,
                "success_rate": rate,
            }
        )

    summary = []
    for packet_loss, rates in sorted(grouped.items()):
        summary.append(
            {
                "packet_loss": packet_loss,
                "seeds": len(rates),
                "mean_success_rate": statistics.mean(rates),
                "variance": statistics.variance(rates) if len(rates) > 1 else 0.0,
                "stddev": statistics.stdev(rates) if len(rates) > 1 else 0.0,
            }
        )

    return rows, summary


def write_csv(rows, summary, output):
    with open(output / "lookup_measurements.csv", "w", newline="", encoding="utf-8") as stream:
        writer = csv.DictWriter(stream, fieldnames=rows[0].keys() if rows else [])
        writer.writeheader()
        writer.writerows(rows)

    with open(output / "packet_loss_summary.csv", "w", newline="", encoding="utf-8") as stream:
        fields = ["packet_loss", "seeds", "mean_success_rate", "variance", "stddev"]
        writer = csv.DictWriter(stream, fieldnames=fields)
        writer.writeheader()
        writer.writerows(summary)


def plot(summary, output):
    losses = [row["packet_loss"] for row in summary]
    means = [row["mean_success_rate"] for row in summary]
    errors = [row["stddev"] for row in summary]

    plt.figure(figsize=(9, 6))
    plt.errorbar(
        losses,
        means,
        yerr=errors,
        marker="o",
        capsize=4,
        linewidth=2,
        label="Measured success rate (mean +/- std. dev.)",
    )
    plt.plot(
        losses,
        [1.0 - loss for loss in losses],
        linestyle="--",
        linewidth=2,
        label="Independent-packet reference: 1 - p",
    )
    plt.xlabel("Packet loss probability p")
    plt.ylabel("Lookup success rate")
    plt.title("Kademlia Lookup Reliability Under Packet Loss")
    plt.ylim(-0.05, 1.05)
    plt.grid(True, alpha=0.3)
    plt.legend()
    plt.tight_layout()
    plt.savefig(output / "success_rate_vs_packet_loss.png", dpi=200)
    plt.close()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("logfile", type=Path)
    parser.add_argument("--output", type=Path, default=Path("packet_loss_results"))
    args = parser.parse_args()

    args.output.mkdir(parents=True, exist_ok=True)
    measurements = parse_log(args.logfile)
    rows, summary = summarize(measurements)
    write_csv(rows, summary, args.output)
    plot(summary, args.output)


if __name__ == "__main__":
    main()
