export function VoiceOrb() {
  return (
    <div className="relative flex items-center justify-center mb-8">
      <div className="absolute size-[360px] rounded-full border border-primary opacity-10" />
      <div className="absolute size-[300px] rounded-full border border-primary opacity-20" />
      <div className="absolute size-[240px] rounded-full border border-primary opacity-40" />
      <div className="relative z-10 flex size-32 items-center justify-center rounded-full bg-primary shadow-2xl">
        <div className="absolute inset-0 rounded-full bg-gradient-to-tr from-black/20 to-transparent opacity-40" />
      </div>
    </div>
  )
}
