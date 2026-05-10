export function VoiceOrb() {
  return (
    <div className="relative flex items-center justify-center mb-8">
      <div
        className="absolute size-[360px] rounded-full border border-primary animate-orb-ring"
        style={{ animationDelay: "1.6s", opacity: 0 }}
      />
      <div
        className="absolute size-[300px] rounded-full border border-primary animate-orb-ring"
        style={{ animationDelay: "0.8s", opacity: 0 }}
      />
      <div
        className="absolute size-[240px] rounded-full border border-primary animate-orb-ring"
        style={{ animationDelay: "0s", opacity: 0 }}
      />
      <div className="relative z-10 flex size-32 items-center justify-center rounded-full bg-primary shadow-2xl animate-orb-bubble">
        <div className="absolute inset-0 rounded-full bg-gradient-to-tr from-black/20 to-transparent opacity-40" />
      </div>
    </div>
  )
}
