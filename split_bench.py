import sys
with open('final_bench.txt', 'r') as f:
    lines = f.readlines()

v1, v2, v2g = [], [], []

for line in lines:
    if line.startswith('BenchmarkV1_'):
        v1.append(line.replace('BenchmarkV1_', 'Benchmark_'))
    elif line.startswith('BenchmarkV2G_'):
        v2g.append(line.replace('BenchmarkV2G_', 'Benchmark_'))
    elif line.startswith('BenchmarkV2_'):
        v2.append(line.replace('BenchmarkV2_', 'Benchmark_'))
    else:
        v1.append(line)
        v2.append(line)
        v2g.append(line)

with open('v1.txt', 'w') as f: f.writelines(v1)
with open('v2.txt', 'w') as f: f.writelines(v2)
with open('v2g.txt', 'w') as f: f.writelines(v2g)
