#!/usr/bin/env python3

"""
Analyze Kademlia lookup scalability experiments.

Expected log format:

# Experiment nr: 0    nodes=10    seed=1    lookups=10
IP:port lookup_contact_start <target_id>
IP:port lookup_contact_rpc 1 <contact_id>
IP:port lookup_contact_rpc 2 <contact_id>
...
IP:port lookup_contact_end

The script measures the number of actual lookup_contact_rpc events
("probes") performed by each lookup.

Usage:

    python analyze_scalability.py data/scalability.txt

Optional:

    python analyze_scalability.py data/scalability.txt --output results

The output directory will contain:

    lookup_measurements.csv
    scalability_summary.csv
    probes_vs_nodes.png
    probes_vs_log2_nodes.png
"""


import argparse
import csv
import math
import re
import statistics
from collections import defaultdict
from pathlib import Path

import matplotlib.pyplot as plt


# ---------------------------------------------------------------------------
# Regular expressions
# ---------------------------------------------------------------------------

EXPERIMENT_RE = re.compile(
    r"^#\s*Experiment\s+nr:\s*(\d+)"
    r"\s+nodes=(\d+)"
    r"\s+seed=(\d+)"
    r"\s+lookups=(\d+)"
)

LOOKUP_START_RE = re.compile(
    r"^(?P<address>\S+)\s+lookup_contact_start\s+(?P<target>\S+)"
)

LOOKUP_RPC_RE = re.compile(
    r"^(?P<address>\S+)\s+lookup_contact_rpc\s+"
    r"(?P<probe>\d+)\s+(?P<contact>\S+)"
)

LOOKUP_END_RE = re.compile(
    r"^(?P<address>\S+)\s+lookup_contact_end"
)


# ---------------------------------------------------------------------------
# Data structures
# ---------------------------------------------------------------------------

class Lookup:
    def __init__(self, experiment, target_id):
        self.experiment = experiment
        self.target_id = target_id
        self.rpcs = []

    def add_rpc(self, probe_number, contact_id):
        self.rpcs.append(
            {
                "probe_number": probe_number,
                "contact_id": contact_id,
            }
        )

    @property
    def probes(self):
        return len(self.rpcs)


# ---------------------------------------------------------------------------
# Parsing
# ---------------------------------------------------------------------------

def parse_log(filename):
    """
    Parse the experiment log.

    Returns:
        list[Lookup]
    """

    lookups = []

    current_experiment = None
    current_lookup = None

    with open(filename, "r", encoding="utf-8") as f:
        for line_number, raw_line in enumerate(f, start=1):
            line = raw_line.strip()

            if not line:
                continue

            # ---------------------------------------------------------------
            # Experiment header
            # ---------------------------------------------------------------
            match = EXPERIMENT_RE.match(line)

            if match:
                # If a lookup was not properly closed, keep it but warn.
                if current_lookup is not None:
                    print(
                        f"WARNING: lookup started before line {line_number} "
                        f"was not closed before a new experiment."
                    )
                    lookups.append(current_lookup)
                    current_lookup = None

                current_experiment = {
                    "experiment": int(match.group(1)),
                    "nodes": int(match.group(2)),
                    "seed": int(match.group(3)),
                    "expected_lookups": int(match.group(4)),
                }

                continue

            # ---------------------------------------------------------------
            # Lookup start
            # ---------------------------------------------------------------
            match = LOOKUP_START_RE.match(line)

            if match:
                if current_experiment is None:
                    print(
                        f"WARNING: lookup before experiment header "
                        f"at line {line_number}"
                    )
                    continue

                if current_lookup is not None:
                    print(
                        f"WARNING: lookup started before previous lookup "
                        f"ended at line {line_number}. "
                        f"Saving previous lookup."
                    )
                    lookups.append(current_lookup)

                current_lookup = Lookup(
                    experiment=current_experiment.copy(),
                    target_id=match.group("target"),
                )

                continue

            # ---------------------------------------------------------------
            # RPC / probe
            # ---------------------------------------------------------------
            match = LOOKUP_RPC_RE.match(line)

            if match:
                if current_lookup is None:
                    print(
                        f"WARNING: RPC without lookup at line "
                        f"{line_number}"
                    )
                    continue

                current_lookup.add_rpc(
                    probe_number=int(match.group("probe")),
                    contact_id=match.group("contact"),
                )

                continue

            # ---------------------------------------------------------------
            # Lookup end
            # ---------------------------------------------------------------
            match = LOOKUP_END_RE.match(line)

            if match:
                if current_lookup is None:
                    print(
                        f"WARNING: lookup end without lookup start "
                        f"at line {line_number}"
                    )
                    continue

                lookups.append(current_lookup)
                current_lookup = None
                continue

            # ---------------------------------------------------------------
            # Unknown line
            # ---------------------------------------------------------------
            if not line.startswith("#"):
                print(
                    f"WARNING: unrecognized line {line_number}: "
                    f"{line}"
                )

    # Handle an unterminated lookup at EOF.
    if current_lookup is not None:
        print("WARNING: final lookup was not terminated with lookup_contact_end.")
        lookups.append(current_lookup)

    return lookups


# ---------------------------------------------------------------------------
# Validation
# ---------------------------------------------------------------------------

def validate_lookups(lookups):
    """
    Validate the parsed measurements.

    Returns a list of warnings.
    """

    warnings = []

    experiments = defaultdict(list)

    for lookup in lookups:
        key = (
            lookup.experiment["experiment"],
            lookup.experiment["nodes"],
            lookup.experiment["seed"],
        )
        experiments[key].append(lookup)

        # A normal lookup should have at least one RPC.
        if lookup.probes == 0:
            warnings.append(
                f"Experiment {key}: lookup {lookup.target_id} "
                f"contains zero RPCs."
            )

        # The probe numbers should normally cover 1..number_of_probes.
        probe_numbers = [
            rpc["probe_number"]
            for rpc in lookup.rpcs
        ]

        expected = set(range(1, len(probe_numbers) + 1))
        actual = set(probe_numbers)

        if expected != actual:
            warnings.append(
                f"Experiment {key}: lookup {lookup.target_id} "
                f"has probe numbers {probe_numbers}; "
                f"expected a permutation of 1..{len(probe_numbers)}."
            )

    # Check expected number of lookups.
    for key, experiment_lookups in experiments.items():
        expected = experiment_lookups[0].experiment["expected_lookups"]

        if len(experiment_lookups) != expected:
            warnings.append(
                f"Experiment {key}: expected {expected} lookups, "
                f"parsed {len(experiment_lookups)}."
            )

    return warnings


# ---------------------------------------------------------------------------
# Per-lookup CSV
# ---------------------------------------------------------------------------

def write_lookup_csv(lookups, filename):
    with open(filename, "w", newline="", encoding="utf-8") as f:
        writer = csv.writer(f)

        writer.writerow(
            [
                "experiment",
                "nodes",
                "seed",
                "target_id",
                "probes",
            ]
        )

        for lookup in lookups:
            exp = lookup.experiment

            writer.writerow(
                [
                    exp["experiment"],
                    exp["nodes"],
                    exp["seed"],
                    lookup.target_id,
                    lookup.probes,
                ]
            )


# ---------------------------------------------------------------------------
# Statistics
# ---------------------------------------------------------------------------

def expected_hops(nodes):
    """
    Approximate expected Kademlia lookup hop count.

    Kademlia uses XOR distances. With random IDs, each successful
    routing step typically reduces the remaining distance by a
    significant amount. A simple asymptotic reference is log2(N).

    We use ceil(log2(N)) as the plotted theoretical reference.
    """

    if nodes <= 1:
        return 0

    return math.ceil(math.log2(nodes))


def calculate_summary(lookups):
    """
    Calculate statistics for each (experiment, nodes) configuration,
    aggregating across the five seeds.

    Returns a list of dictionaries.
    """

    # First calculate statistics independently for each seed.
    per_seed = defaultdict(list)

    for lookup in lookups:
        exp = lookup.experiment

        key = (
            exp["experiment"],
            exp["nodes"],
            exp["seed"],
        )

        per_seed[key].append(lookup.probes)

    # Then aggregate the individual lookup measurements across seeds.
    by_nodes = defaultdict(list)

    for (experiment, nodes, seed), probes in per_seed.items():
        by_nodes[(experiment, nodes)].extend(probes)

    summary = []

    for (experiment, nodes), probes in sorted(by_nodes.items()):
        mean = statistics.mean(probes)

        if len(probes) >= 2:
            variance = statistics.variance(probes)
            stddev = statistics.stdev(probes)
        else:
            variance = 0.0
            stddev = 0.0

        summary.append(
            {
                "experiment": experiment,
                "nodes": nodes,
                "num_lookups": len(probes),
                "mean_probes": mean,
                "variance": variance,
                "stddev": stddev,
                "min_probes": min(probes),
                "max_probes": max(probes),
                "expected_log2_hops": expected_hops(nodes),
            }
        )

    return summary, per_seed


def write_summary_csv(summary, filename):
    with open(filename, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(
            f,
            fieldnames=[
                "experiment",
                "nodes",
                "num_lookups",
                "mean_probes",
                "variance",
                "stddev",
                "min_probes",
                "max_probes",
                "expected_log2_hops",
            ],
        )

        writer.writeheader()
        writer.writerows(summary)


# ---------------------------------------------------------------------------
# Plotting
# ---------------------------------------------------------------------------

def plot_probes_vs_nodes(summary, output_file):
    """
    Plot average probes against network size.

    Error bars show one standard deviation.
    The dashed reference curve is ceil(log2(N)).
    """

    nodes = [row["nodes"] for row in summary]
    means = [row["mean_probes"] for row in summary]
    stddevs = [row["stddev"] for row in summary]
    expected = [row["expected_log2_hops"] for row in summary]

    plt.figure(figsize=(9, 6))

    plt.errorbar(
        nodes,
        means,
        yerr=stddevs,
        marker="o",
        capsize=4,
        linewidth=2,
        label="Measured probes (mean ± std. dev.)",
    )

    plt.plot(
        nodes,
        expected,
        marker="x",
        linestyle="--",
        linewidth=2,
        label=r"Reference: $\lceil\log_2(N)\rceil$",
    )

    plt.xscale("log", base=2)

    plt.xlabel("Network size N (nodes)")
    plt.ylabel("Number of lookup probes")
    plt.title("Kademlia Lookup Scalability")
    plt.grid(True, alpha=0.3)
    plt.legend()
    plt.tight_layout()

    plt.savefig(output_file, dpi=200)
    plt.close()


def plot_probes_vs_log2_nodes(summary, output_file):
    """
    Plot average probes against log2(N).

    This makes the expected logarithmic relationship visually easier
    to inspect.
    """

    x = [
        math.log2(row["nodes"])
        for row in summary
    ]

    means = [
        row["mean_probes"]
        for row in summary
    ]

    stddevs = [
        row["stddev"]
        for row in summary
    ]

    plt.figure(figsize=(9, 6))

    plt.errorbar(
        x,
        means,
        yerr=stddevs,
        marker="o",
        capsize=4,
        linewidth=2,
    )

    plt.xlabel(r"$\log_2(N)$")
    plt.ylabel("Number of lookup probes")
    plt.title("Lookup Probes vs. log₂(Network Size)")
    plt.grid(True, alpha=0.3)
    plt.tight_layout()

    plt.savefig(output_file, dpi=200)
    plt.close()


# ---------------------------------------------------------------------------
# Console report
# ---------------------------------------------------------------------------

def print_report(summary, per_seed):
    print()
    print("=" * 75)
    print("Kademlia lookup scalability")
    print("=" * 75)

    print(
        f"{'N':>8} "
        f"{'lookups':>10} "
        f"{'mean':>12} "
        f"{'stddev':>12} "
        f"{'variance':>12} "
        f"{'min':>8} "
        f"{'max':>8} "
        f"{'log2(N)':>10}"
    )

    print("-" * 75)

    for row in summary:
        print(
            f"{row['nodes']:>8} "
            f"{row['num_lookups']:>10} "
            f"{row['mean_probes']:>12.3f} "
            f"{row['stddev']:>12.3f} "
            f"{row['variance']:>12.3f} "
            f"{row['min_probes']:>8} "
            f"{row['max_probes']:>8} "
            f"{row['expected_log2_hops']:>10}"
        )

    print()
    print("Per-seed means")
    print("-" * 75)

    print(
        f"{'N':>8} "
        f"{'seed':>8} "
        f"{'lookups':>10} "
        f"{'mean probes':>15}"
    )

    print("-" * 50)

    for key in sorted(per_seed):
        experiment, nodes, seed = key
        probes = per_seed[key]

        print(
            f"{nodes:>8} "
            f"{seed:>8} "
            f"{len(probes):>10} "
            f"{statistics.mean(probes):>15.3f}"
        )


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser(
        description="Analyze Kademlia scalability experiment logs."
    )

    parser.add_argument(
        "logfile",
        help="Path to scalability log file",
    )

    parser.add_argument(
        "--output",
        default="scalability_results",
        help="Output directory (default: scalability_results)",
    )

    args = parser.parse_args()

    logfile = Path(args.logfile)
    output_dir = Path(args.output)

    output_dir.mkdir(parents=True, exist_ok=True)

    print(f"Reading {logfile}...")

    lookups = parse_log(logfile)

    if not lookups:
        raise RuntimeError("No lookup measurements found.")

    print(f"Parsed {len(lookups)} lookups.")

    # Validate the input.
    warnings = validate_lookups(lookups)

    if warnings:
        print()
        print("Warnings:")
        for warning in warnings:
            print("  -", warning)

    # Calculate statistics.
    summary, per_seed = calculate_summary(lookups)

    # Write CSV files.
    lookup_csv = output_dir / "lookup_measurements.csv"
    summary_csv = output_dir / "scalability_summary.csv"

    write_lookup_csv(lookups, lookup_csv)
    write_summary_csv(summary, summary_csv)

    # Generate plots.
    plot_probes_vs_nodes(
        summary,
        output_dir / "probes_vs_nodes.png",
    )

    plot_probes_vs_log2_nodes(
        summary,
        output_dir / "probes_vs_log2_nodes.png",
    )

    # Print report.
    print_report(summary, per_seed)

    print()
    print("=" * 75)
    print("Output")
    print("=" * 75)
    print(f"Individual measurements: {lookup_csv}")
    print(f"Summary:                 {summary_csv}")
    print(
        f"Plot:                    "
        f"{output_dir / 'probes_vs_nodes.png'}"
    )
    print(
        f"Plot:                    "
        f"{output_dir / 'probes_vs_log2_nodes.png'}"
    )


if __name__ == "__main__":
    main()