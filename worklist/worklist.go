package worklist

//keeps track of files to be processed

//entry in the channel
type Entry struct {
	Path string
}

type Worklist struct {
	jobs chan Entry
}

//add a job to the channel
func (w *Worklist) Add(work Entry) {
	w.jobs <- work
}

//get a job off the channel
func (w *Worklist) Next() Entry {
	j := <-w.jobs
	return j
}

//generate a new worklist (buffered channel)
func New(bufSize int) Worklist {
	return Worklist{make(chan Entry, bufSize)}
}

//create new job entry
func NewJob(path string) Entry {
	return Entry{path}
}

//generate empty jobs - signal to workers that they should quit
func (w *Worklist) Finalize(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		emptyJob := NewJob("")
		w.Add(emptyJob)
	}
}
