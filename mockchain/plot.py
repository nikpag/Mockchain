import matplotlib
import matplotlib.pyplot as plt
import numpy as np

matplotlib.use("Agg")

data = {
    "5nodes": {
        "D4": {
            "C1":{
                "throughput": 0.61703,
                "blocktime": 1.62046,
            },
            "C5":{
                "throughput": 0.661203828,
                "blocktime": 7.56592526,
            },
            "C10":{
                "throughput": 0.732586,
                "blocktime": 13.6926268,
            }
        },
        "D5": {
            "C1":{
                "throughput": 0.0575,
                "blocktime": 17.3806,
            },
            "C5":{
                "throughput": 0.049,
                "blocktime": 103,
            },
            "C10":{
                "throughput": 0.064,
                "blocktime": 156.25,
            }
        },
    },
    "10nodes": {
        "D4": {
            "C1":{
                "throughput": 0.75,
                "blocktime": 1.333,
            },
            "C5":{
                "throughput": 0.318,
                "blocktime": 5.723,
            },
            "C10":{
                "throughput": 0.46242774,
                "blocktime": 21.625,
            }
        },
        "D5": {
            "C1":{
                "throughput": 0.07797271,
                "blocktime": 12.825,
            },
            "C5":{
                "throughput": 0.058,
                "blocktime": 84,
            },
            "C10":{
                "throughput": 0.056,
                "blocktime": 138.84,
            }
        },
    }
}

throughput5 = [data["5nodes"][d][c]["throughput"] for d in data["5nodes"] for c in data["5nodes"][d]]
blocktime5 = [data["5nodes"][d][c]["blocktime"] for d in data["5nodes"] for c in data["5nodes"][d]]
throughput10 = [data["10nodes"][d][c]["throughput"] for d in data["10nodes"] for c in data["10nodes"][d]]
blocktime10 = [data["10nodes"][d][c]["blocktime"] for d in data["10nodes"] for c in data["10nodes"][d]]

print(throughput5)
print(blocktime5)
print(throughput10)
print(blocktime10)

fig, ax1 = plt.subplots()
ax2 = ax1.twinx()

marker="o-"

ax1.plot(np.arange(len(throughput5)), throughput5, marker, color="lightgreen", label="Throughput: 5 nodes")
ax1.plot(np.arange(len(throughput10)), throughput10, marker, color="darkgreen", label="Throughput: 10 nodes")

ax1.set_xlabel("Difficulty.Capacity")
ax1.set_ylabel("Throughput (Transactions/sec)")
ax1.set_xticks(np.arange(len(throughput5)))
ax1.set_xticklabels([f"{k1}.{k2}" for k1 in data["5nodes"] for k2 in data["5nodes"][k1]])
ax1.grid(True)

ax2.plot(np.arange(len(blocktime5)), blocktime5, marker, color="red", label="Block time: 5 nodes")
ax2.plot(np.arange(len(blocktime10)), blocktime10, marker, color="darkred", label="Block time: 10 nodes")

ax2.set_ylabel("Average block time (sec)")

fig.legend(bbox_to_anchor=(1.3, 1))

plt.title("Transaction throughput & Average block time for 5/10 nodes")
plt.savefig("plot.pdf", bbox_inches="tight")
