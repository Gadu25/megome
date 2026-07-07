let SessionLoad = 1
let s:so_save = &g:so | let s:siso_save = &g:siso | setg so=0 siso=0 | setl so=-1 siso=-1
let v:this_session=expand("<sfile>:p")
doautoall SessionLoadPre
silent only
silent tabonly
cd ~/my-go/megome
if expand('%') == '' && !&modified && line('$') <= 1 && getline(1) == ''
  let s:wipebuf = bufnr('%')
endif
let s:shortmess_save = &shortmess
set shortmess+=aoO
badd +1 ~/my-go
badd +49 term://~/my-go//7514:/bin/bash
badd +1 .env.backup
badd +210 term://~/my-go/megome//7783:/bin/bash
badd +1 internal/services/mailer/renderer.go
badd +103 cmd/api/api.go
badd +477 internal/services/user/routes.go
badd +1 internal/services/mailer/templates/base.html
badd +1 internal/services/mailer/templates/reset_password.html
badd +1 internal/services/mailer/service.go
badd +50 cmd/main.go
badd +51 internal/services/mailer/mailer.go
badd +16 internal/services/types/types.go
badd +22 internal/services/user/store.go
badd +3 ~/my-go/megome-front/app/page.tsx
badd +1 ~/my-go/megome-front/components/form/Profile.tsx
badd +1 ~/my-go/megome-front/types/types.ts
badd +1 ~/my-go/megome-front/lib/api/client/profile.ts
badd +1 ~/my-go/megome-front/features/profile/schema.ts
badd +256 ~/my-go/megome-front/components/profile/TopProfile.tsx
badd +128 internal/services/profile/store.go
badd +25 internal/services/profile/routes.go
argglobal
%argdel
$argadd ~/my-go
set stal=2
tabnew +setlocal\ bufhidden=wipe
tabrewind
edit ~/my-go/megome-front/components/form/Profile.tsx
let s:save_splitbelow = &splitbelow
let s:save_splitright = &splitright
set splitbelow splitright
wincmd _ | wincmd |
vsplit
1wincmd h
wincmd w
let &splitbelow = s:save_splitbelow
let &splitright = s:save_splitright
wincmd t
let s:save_winminheight = &winminheight
let s:save_winminwidth = &winminwidth
set winminheight=0
set winheight=1
set winminwidth=0
set winwidth=1
exe 'vert 1resize ' . ((&columns * 93 + 94) / 188)
exe 'vert 2resize ' . ((&columns * 94 + 94) / 188)
argglobal
balt ~/my-go/megome-front/components/profile/TopProfile.tsx
setlocal foldmethod=manual
setlocal foldexpr=0
setlocal foldmarker={{{,}}}
setlocal foldignore=#
setlocal foldlevel=0
setlocal foldminlines=1
setlocal foldnestmax=20
setlocal foldenable
silent! normal! zE
let &fdl = &fdl
let s:l = 46 - ((35 * winheight(0) + 26) / 53)
if s:l < 1 | let s:l = 1 | endif
keepjumps exe s:l
normal! zt
keepjumps 46
normal! 033|
wincmd w
argglobal
if bufexists(fnamemodify("~/my-go/megome-front/types/types.ts", ":p")) | buffer ~/my-go/megome-front/types/types.ts | else | edit ~/my-go/megome-front/types/types.ts | endif
if &buftype ==# 'terminal'
  silent file ~/my-go/megome-front/types/types.ts
endif
balt internal/services/profile/store.go
setlocal foldmethod=manual
setlocal foldexpr=0
setlocal foldmarker={{{,}}}
setlocal foldignore=#
setlocal foldlevel=0
setlocal foldminlines=1
setlocal foldnestmax=20
setlocal foldenable
silent! normal! zE
let &fdl = &fdl
let s:l = 25 - ((24 * winheight(0) + 26) / 53)
if s:l < 1 | let s:l = 1 | endif
keepjumps exe s:l
normal! zt
keepjumps 25
normal! 014|
lcd ~/my-go
wincmd w
exe 'vert 1resize ' . ((&columns * 93 + 94) / 188)
exe 'vert 2resize ' . ((&columns * 94 + 94) / 188)
tabnext
let s:save_splitbelow = &splitbelow
let s:save_splitright = &splitright
set splitbelow splitright
wincmd _ | wincmd |
vsplit
1wincmd h
wincmd w
let &splitbelow = s:save_splitbelow
let &splitright = s:save_splitright
wincmd t
let s:save_winminheight = &winminheight
let s:save_winminwidth = &winminwidth
set winminheight=0
set winheight=1
set winminwidth=0
set winwidth=1
exe 'vert 1resize ' . ((&columns * 93 + 94) / 188)
exe 'vert 2resize ' . ((&columns * 94 + 94) / 188)
argglobal
if bufexists(fnamemodify("term://~/my-go//7514:/bin/bash", ":p")) | buffer term://~/my-go//7514:/bin/bash | else | edit term://~/my-go//7514:/bin/bash | endif
if &buftype ==# 'terminal'
  silent file term://~/my-go//7514:/bin/bash
endif
setlocal foldmethod=manual
setlocal foldexpr=0
setlocal foldmarker={{{,}}}
setlocal foldignore=#
setlocal foldlevel=0
setlocal foldminlines=1
setlocal foldnestmax=20
setlocal foldenable
let s:l = 49 - ((44 * winheight(0) + 26) / 53)
if s:l < 1 | let s:l = 1 | endif
keepjumps exe s:l
normal! zt
keepjumps 49
normal! 035|
wincmd w
argglobal
if bufexists(fnamemodify("term://~/my-go/megome//7783:/bin/bash", ":p")) | buffer term://~/my-go/megome//7783:/bin/bash | else | edit term://~/my-go/megome//7783:/bin/bash | endif
if &buftype ==# 'terminal'
  silent file term://~/my-go/megome//7783:/bin/bash
endif
balt term://~/my-go//7514:/bin/bash
setlocal foldmethod=manual
setlocal foldexpr=0
setlocal foldmarker={{{,}}}
setlocal foldignore=#
setlocal foldlevel=0
setlocal foldminlines=1
setlocal foldnestmax=20
setlocal foldenable
let s:l = 209 - ((51 * winheight(0) + 26) / 53)
if s:l < 1 | let s:l = 1 | endif
keepjumps exe s:l
normal! zt
keepjumps 209
normal! 0
wincmd w
exe 'vert 1resize ' . ((&columns * 93 + 94) / 188)
exe 'vert 2resize ' . ((&columns * 94 + 94) / 188)
tabnext 2
set stal=1
if exists('s:wipebuf') && len(win_findbuf(s:wipebuf)) == 0 && getbufvar(s:wipebuf, '&buftype') isnot# 'terminal'
  silent exe 'bwipe ' . s:wipebuf
endif
unlet! s:wipebuf
set winheight=1 winwidth=20
let &shortmess = s:shortmess_save
let &winminheight = s:save_winminheight
let &winminwidth = s:save_winminwidth
let s:sx = expand("<sfile>:p:r")."x.vim"
if filereadable(s:sx)
  exe "source " . fnameescape(s:sx)
endif
let &g:so = s:so_save | let &g:siso = s:siso_save
set hlsearch
nohlsearch
doautoall SessionLoadPost
unlet SessionLoad
" vim: set ft=vim :
