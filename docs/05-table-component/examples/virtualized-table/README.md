# Data Virtualization Example

A large employee database demonstrating VTable's data virtualization capabilities with 10,000 employees loaded efficiently through chunk-based loading.

## Features Demonstrated

- **Large dataset handling** with 10,000+ simulated employees
- **Automatic chunk loading** with configurable chunk sizes (25 rows per chunk)
- **Loading state feedback** showing chunk loading progress
- **Memory efficiency** using only ~50KB regardless of dataset size
- **Smooth scrolling performance** with predictive loading
- **Smart caching** with automatic memory management
- **Realistic data simulation** with random departments, statuses, and salaries

## Data Virtualization Benefits

### Memory Usage
- **10,000 rows**: Uses only ~50KB memory
- **100,000 rows**: Still ~50KB memory
- **1,000,000 rows**: Still ~50KB memory!

### Performance
- **Immediate response**: Scrolling never blocks
- **Predictive loading**: Chunks load ahead of scrolling
- **Automatic cleanup**: Unloads distant chunks to free memory

## Running the Example

```bash
cd docs/05-table-component/examples/virtualized-table
go run .
```

## Controls

| Key | Action |
|-----|--------|
| `↑` `k` | Move up one row |
| `↓` `j` | Move down one row |
| `g` | Jump to first employee |
| `G` | Jump to last employee |
| `h` `PgUp` | Jump up 10 rows |
| `l` `PgDn` | Jump down 10 rows |
| `J` | Open jump-to-index form |
| `q` | Quit |

## What You'll See

### Normal Operation
```
✅ Large Employee Database (10000 employees) | Position: 1/10000 | Use ↑↓ j/k g/G PgUp/PgDn, q to quit

│ ●  │Employee Name       │  Department   │   Status   │      Salary│
│ ►  │Employee 1          │  Engineering  │   Active   │     $67,000│
│    │Employee 2          │   Marketing   │   Remote   │     $58,000│  
│    │Employee 3          │      Sales    │   Active   │     $73,000│
│    │Employee 4          │        HR     │  On Leave  │     $51,000│
│    │Employee 5          │  Engineering  │   Active   │     $89,000│
│    │Employee 6          │   Marketing   │   Remote   │     $64,000│
│    │Employee 7          │     Finance   │   Active   │     $76,000│
│    │Employee 8          │  Operations   │   Active   │     $52,000│
│    │Employee 9          │      Sales    │   Active   │     $68,000│
│    │Employee 10         │  Engineering  │  On Leave  │     $91,000│
```

### During Loading
```
Loading 2 chunks... | Position: 847/10000 | Use ↑↓ j/k g/G PgUp/PgDn, q to quit
```

## Configuration Options

### Viewport Configuration
```go
ViewportConfig: core.ViewportConfig{
    Height:             10, // Visible rows
    ChunkSize:          25, // Rows per chunk
    TopThreshold:       3,  // Load trigger (scrolling up)
    BottomThreshold:    3,  // Load trigger (scrolling down)
    BoundingAreaBefore: 50, // Memory buffer before viewport
    BoundingAreaAfter:  50, // Memory buffer after viewport
}
```

### Tuning Performance

**For slower networks/databases:**
- Increase `ChunkSize` to 50+ (fewer requests)
- Increase thresholds to 5+ (more aggressive loading)

**For faster networks/databases:**
- Decrease `ChunkSize` to 10-15 (more responsive)
- Decrease thresholds to 1-2 (just-in-time loading)

**For limited memory:**
- Decrease `BoundingAreaBefore/After` to 25 (smaller cache)
- Decrease `ChunkSize` to 15-20 (smaller chunks)

## Try These Experiments

### 1. Dataset Size
Change the employee count in `main()`:
```go
totalEmployees := 100000  // 100k employees
totalEmployees := 1000000 // 1M employees  
```

Notice that memory usage and initial performance remain the same!

### 2. Chunk Size
Modify the config:
```go
ChunkSize: 50,  // Larger chunks = fewer requests
ChunkSize: 10,  // Smaller chunks = more responsive
```

### 3. Loading Delay
Adjust the simulated database delay:
```go
// In LoadChunk method
time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond) // Slower "database"
time.Sleep(time.Duration(10+rand.Intn(20)) * time.Millisecond)   // Faster "database"
```

### 4. Scroll Behavior  
Try different scrolling patterns:
- **Fast scrolling**: Hold `l` or Page Down and watch chunk loading
- **Jump to end**: Press `G` and see how VTable handles large jumps
- **Random access**: Jump around with `g`, `G`, `h/l`, `PgUp/PgDn`

## Real-World Applications

This pattern works well for:

- **Employee directories** (HR systems)
- **Customer databases** (CRM systems)  
- **Product catalogs** (e-commerce)
- **Log viewers** (system monitoring)
- **Financial records** (accounting systems)
- **Any large tabular dataset** where loading everything at once would be slow or memory-intensive

## Implementation Notes

### Data Generation
- **Realistic simulation**: Random departments, statuses, salaries
- **Unique IDs**: Each employee has a stable ID for selection tracking
- **On-demand generation**: Data created only when requested (simulates database queries)

### Loading Simulation
- **Variable delays**: 100-300ms to simulate real database response times
- **Predictable data**: Same employee ID always generates same data (simulates consistent database)
- **Memory efficiency**: No data stored permanently, regenerated as needed

This example demonstrates how VTable can make working with large datasets feel as responsive as working with small ones. 