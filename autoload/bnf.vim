"   filename: bnf.vim

if get(s:, 'loaded', 0)
    finish
endif

let s:loaded = 1
let g:bnf#source = extend(get(g:, 'bnf#source', {}), {
            \ 'name': 'nvim-bnf',
            \ 'ready': 1,
            \ 'priority': 9,
            \ 'mark': 'bnf',
            \ 'scope': ['bnf'],
            \ 'complete_pattern': '<',
            \ 'on_complete': 'bnf#on_complete',
            \ 'on_warmup': 'bnf#on_warmup',
            \ }, 'keep')

func! bnf#init()
    call ncm2#register_source(g:bnf#source)
endfunc

func! bnf#on_warmup(ctx)
    call BNFNcm2OnWarmup(a:ctx)
endfunc

func! bnf#on_complete(ctx)
    call BNFNcm2OnComplete(a:ctx)
endfunc
