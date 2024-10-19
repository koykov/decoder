package decoder

// node object that describes one operator in decoder's body.
type node struct {
	typ rtype
	// Destination/source pair.
	dst, src, ins []byte
	// List of keys, that need to check sequentially in the source object.
	subset [][]byte
	// Getter callback, for sources like "dst = getFoo(var0, ...)"
	getter GetterFn
	// Callback for lines like "prepareObject(var.obj)"
	callback CallbackFn
	// Flag that indicates if source is a static value.
	static bool
	// List of modifier applied to source.
	mod []mod
	// List of arguments for getter or callback.
	arg []*arg
	// List of children nodes.
	child []node

	// Loop stuff.
	loopKey       []byte
	loopVal       []byte
	loopSrc       []byte
	loopCnt       []byte
	loopCntInit   []byte
	loopCntStatic bool
	loopCntOp     op
	loopCondOp    op
	loopLim       []byte
	loopLimStatic bool
	loopBrkD      int

	// Condition stuff.
	condL, condOKL []byte
	condR, condOKR []byte
	condStaticL    bool
	condStaticR    bool
	condOp         op
	condHlp        []byte
	condHlpArg     []*arg
	condIns        []byte
	condLC         lc
}
