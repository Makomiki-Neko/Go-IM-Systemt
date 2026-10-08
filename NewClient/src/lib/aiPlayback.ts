/** One-second receive windows, then a one-second typewriter playback.
 * Fragments are sorted only; no gap checking, deduplication or retransmission.
 */
export class AIPlayback {
  private pending: {seq:number; content:string; reasoning:string}[] = []
  private window: ReturnType<typeof setInterval>
  private animation?: ReturnType<typeof setInterval>
  private finishFrame?: () => void
  private content = ''
  private reasoning = ''
  constructor(private update:(content:string,reasoning:string)=>void) {
    this.window = setInterval(() => this.play(), 1000)
  }
  push(seq:number,content:string,reasoning:string) { this.pending.push({seq,content,reasoning}) }
  private play() {
    this.finishFrame?.(); clearInterval(this.animation); this.finishFrame = undefined
    const batch=this.pending.splice(0).sort((a,b)=>a.seq-b.seq)
    if (!batch.length) return
    // Keep channel order within each received fragment; reasoning precedes its answer.
    const chars=batch.flatMap(d=>[...Array.from(d.reasoning).map(char=>({char,reasoning:true})),...Array.from(d.content).map(char=>({char,reasoning:false}))])
    let shown=0
    const start=performance.now()
    const reveal=(n:number)=>{while(shown<n){const c=chars[shown++];if(c.reasoning)this.reasoning+=c.char;else this.content+=c.char}this.update(this.content,this.reasoning)}
    this.finishFrame=()=>reveal(chars.length)
    this.animation=setInterval(()=>{reveal(Math.min(chars.length,Math.floor(chars.length*(performance.now()-start)/1000)));if(shown===chars.length){clearInterval(this.animation);this.finishFrame=undefined}},16)
  }
  stop() {clearInterval(this.window);clearInterval(this.animation);this.pending=[];this.finishFrame=undefined}
}
