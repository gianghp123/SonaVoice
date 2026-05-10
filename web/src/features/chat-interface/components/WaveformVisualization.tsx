const BARS = [
  { height: "h-2" },
  { height: "h-4" },
  { height: "h-3" },
  { height: "h-6" },
  { height: "h-8" },
  { height: "h-10" },
  { height: "h-8" },
  { height: "h-5" },
  { height: "h-4" },
  { height: "h-7" },
  { height: "h-3" },
  { height: "h-4" },
]

export function WaveformVisualization() {
  return (
    <div className="flex items-end justify-center h-10 gap-0 mb-8">
      {BARS.map((bar, i) => (
        <div
          key={i}
          className={`w-[3px] mx-[1.5px] rounded-sm bg-primary ${bar.height}`}
        />
      ))}
    </div>
  )
}
