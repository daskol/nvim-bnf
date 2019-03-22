"   filename: bnf.vim

if exists('g:loaded_nvim_bnf')
  finish
else
  let g:loaded_nvim_bnf = 1
endif

function! s:RequireHost(host) abort
  return jobstart(['nvim-bnf'], {'rpc': v:true})
endfunction

" Register tast-specific plugin host and register plugin.
call remote#host#Register('nvim-bnf', 'x', function('s:RequireHost'))
call remote#host#RegisterPlugin('nvim-bnf', '0', [
\ {'type': 'autocmd', 'name': 'BufNewFile', 'sync': 0, 'opts': {'eval': 'expand("<afile>")', 'group': 'nvim-bnf', 'pattern': '*.bnf'}},
\ {'type': 'autocmd', 'name': 'BufRead', 'sync': 0, 'opts': {'eval': 'expand("<afile>")', 'group': 'nvim-bnf', 'pattern': '*.bnf'}},
\ ])
