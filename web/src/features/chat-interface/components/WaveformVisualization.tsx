const BARS = [
  { height: "h-2", delay: "0s" },
  { height: "h-4", delay: "0.1s" },
  { height: "h-3", delay: "0.2s" },
  { height: "h-6", delay: "0.3s" },
  { height: "h-8", delay: "0.4s" },
  { height: "h-10", delay: "0.5s" },
  { height: "h-8", delay: "0.6s" },
  { height: "h-5", delay: "0.7s" },
  { height: "h-4", delay: "0.8s" },
  { height: "h-7", delay: "0.9s" },
  { height: "h-3", delay: "1s" },
  { height: "h-4", delay: "1.1s" },
]

export function WaveformVisualization() {
  return (
    <div className="flex items-end justify-center h-10 gap-0 mb-8">
      {BARS.map((bar, i) => (
        <div
          key={i}
          className={`w-[3px] mx-[1.5px] rounded-sm bg-primary animate-waveform-bar origin-bottom ${bar.height}`}
          style={{ animationDelay: bar.delay }}
        />
      ))}
    </div>
  )
}
