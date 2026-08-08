// Slate Setup wizard — install detection, audit tool, updates, project protection.

const STEPS = ['Home', 'Check PC', 'Install', 'Finish']

function el(tag, attrs = {}, children = []) {
  const n = document.createElement(tag)
  for (const [k, v] of Object.entries(attrs)) {
    if (k === 'className') n.className = v
    else if (k === 'text') n.textContent = v
    else if (k.startsWith('on') && typeof v === 'function') n.addEventListener(k.slice(2).toLowerCase(), v)
    else if (k === 'checked') n.checked = !!v
    else if (v !== false && v != null) n.setAttribute(k, v)
  }
  for (const c of [].concat(children)) {
    if (c == null) continue
    n.appendChild(typeof c === 'string' ? document.createTextNode(c) : c)
  }
  return n
}

const state = {
  // home | audit | wizard-check | wizard-install | finish | deep-audit
  view: 'home',
  step: 0,
  audit: null,
  deep: null,
  installStatus: null,
  updates: null,
  desktop: true,
  busy: false,
  progress: { step: '', detail: '', percent: 0 },
  result: null,
  brain: null,
  error: null,
  paths: {},
  mode: 'install' // install | update | repair
}

function go() {
  return window.go?.main?.App
}

function eventsOn(name, cb) {
  if (window.runtime?.EventsOn) return window.runtime.EventsOn(name, cb)
  return () => {}
}

async function bootstrap() {
  try {
    state.paths = (await go().GetPaths()) || {}
    state.installStatus = await go().GetInstallStatus()
  } catch (e) {
    console.warn(e)
  }
  render()
}

function render() {
  const root = document.getElementById('app')
  root.innerHTML = ''

  root.appendChild(
    el('div', { className: 'header' }, [
      el('h1', { text: '◆  Slate Setup' }),
      el('p', {
        text: 'Helper to install Sam Wasserman\'s Slate on Windows. Your project files are always protected.'
      }),
      el('p', { className: 'credit' }, [
        'Slate by Sam Wasserman (Apache-2.0). This Setup program is a Windows install helper only — not by Sam. ',
        el('a', {
          href: '#',
          onClick: (e) => {
            e.preventDefault()
            go()?.OpenExternal('https://github.com/wassermanproductions/slate')
          }
        }, ['Slate on GitHub'])
      ])
    ])
  )

  if (state.view === 'home' || state.view === 'deep-audit') {
    // no step pills
  } else {
    const pills = el('div', { className: 'steps' })
    STEPS.forEach((name, i) => {
      const cls = 'step-pill' + (i === state.step ? ' on' : '') + (i < state.step ? ' done' : '')
      pills.appendChild(el('div', { className: cls, text: `${i + 1}. ${name}` }))
    })
    root.appendChild(pills)
  }

  const main = el('div', { className: 'main' })
  if (state.view === 'home') main.appendChild(viewHome())
  else if (state.view === 'deep-audit') main.appendChild(viewDeepAudit())
  else if (state.step === 1) main.appendChild(viewAudit())
  else if (state.step === 2) main.appendChild(viewInstall())
  else if (state.step === 3) main.appendChild(viewFinish())
  else main.appendChild(viewHome())
  root.appendChild(main)

  root.appendChild(viewFooter())
}

function statusBanner() {
  const st = state.installStatus
  if (!st) return el('div', { className: 'card' }, [el('p', { text: 'Checking install status…' })])

  const cls = st.installHealthy ? 'okbox' : st.installed ? 'summary' : 'summary'
  const lines = [st.summary, ...(st.detailLines || [])].join('\n')
  const title =
    st.flavor === 'npm'
      ? 'Your PC — Slate found (npm / Electron)'
      : st.installed
        ? 'Your PC — Slate status'
        : 'Your PC — Slate status'
  return el('div', { className: 'card' }, [
    el('h2', { text: title }),
    el('div', { className: cls, text: lines })
  ])
}

function viewHome() {
  const wrap = el('div', {})
  wrap.appendChild(statusBanner())

  if (state.updates) {
    const u = state.updates
    wrap.appendChild(
      el('div', { className: 'card' }, [
        el('h2', { text: 'Repository updates' }),
        el('div', {
          className: u.updateAvailable ? 'summary' : u.upToDate ? 'okbox' : 'summary',
          text: [u.summary, u.detail, u.projectsNote].filter(Boolean).join('\n\n')
        })
      ])
    )
  }

  wrap.appendChild(
    el('div', { className: 'card' }, [
      el('h2', { text: 'What do you want to do?' }),
      el('p', {
        text: 'Choose one. You can run the Audit tool anytime without changing anything on your PC.'
      }),
      el('div', { className: 'btn-row', style: 'margin-top:12px;flex-direction:column;align-items:stretch' }, [
        btnPrimary(
          state.installStatus?.installed ? 'Repair / reinstall Slate' : 'Install Slate',
          () => startWizard('install')
        ),
        state.updates?.updateAvailable
          ? btnPrimary('Install update (keeps your projects)', () => startWizard('update'))
          : null,
        el('button', {
          text: 'Audit tool — full system & install check',
          onClick: () => runDeepAudit()
        }),
        el('button', {
          text: 'Check for updates from GitHub',
          disabled: state.busy,
          onClick: () => checkUpdates()
        }),
        state.installStatus?.installed && state.installStatus?.flavor === 'packaged'
          ? el('button', {
              text: 'Open Slate',
              onClick: () => go()?.LaunchSlateProcess()
            })
          : null,
        state.installStatus?.installed && state.installStatus?.flavor === 'npm'
          ? el('button', {
              text: 'Open Slate folder',
              onClick: () => {
                const root = state.installStatus.installDir
                if (root) go()?.OpenExternal('file:///' + root.replace(/\\/g, '/'))
              }
            })
          : null,
        el('button', {
          className: 'ghost',
          text: 'Open my projects folder (safe)',
          onClick: () => go()?.OpenProjectsFolder()
        })
      ])
    ])
  )

  if (state.error) wrap.appendChild(el('div', { className: 'err', text: state.error }))
  return wrap
}

function btnPrimary(text, onClick) {
  return el('button', { className: 'primary', text, onClick, disabled: state.busy })
}

function viewDeepAudit() {
  const card = el('div', { className: 'card' }, [el('h2', { text: 'Audit tool (read-only)' })])
  if (!state.deep) {
    card.appendChild(el('p', { text: 'Running audit…' }))
    return card
  }
  const d = state.deep
  card.appendChild(el('div', { className: 'summary', text: d.summary || '' }))
  const ul = el('ul', {})
  for (const h of d.highlights || []) ul.appendChild(el('li', { text: h }))
  card.appendChild(ul)

  // System checks
  card.appendChild(el('h2', { text: 'System checks', style: 'margin-top:16px' }))
  const list = el('ul', { className: 'check-list' })
  for (const c of d.system?.checks || []) {
    list.appendChild(
      el('li', {}, [
        el('div', { className: 'dot ' + (c.ok ? 'ok' : c.required ? 'bad' : 'warn') }),
        el('div', {}, [
          el('div', { className: 'check-title', text: c.label }),
          el('div', { className: 'check-detail', text: c.detail }),
          el('div', { className: 'check-detail', text: c.action })
        ])
      ])
    )
  }
  card.appendChild(list)

  // Install block
  if (d.install) {
    card.appendChild(el('h2', { text: 'Install health', style: 'margin-top:16px' }))
    card.appendChild(
      el('div', {
        className: d.install.installHealthy ? 'okbox' : 'summary',
        text: (d.install.detailLines || []).join('\n')
      })
    )
  }

  // Updates
  if (d.updates) {
    card.appendChild(el('h2', { text: 'GitHub updates', style: 'margin-top:16px' }))
    card.appendChild(
      el('div', {
        className: d.updates.updateAvailable ? 'summary' : 'okbox',
        text: [d.updates.summary, d.updates.detail, d.updates.projectsNote].filter(Boolean).join('\n\n')
      })
    )
  }

  if (state.error) card.appendChild(el('div', { className: 'err', text: state.error }))
  return card
}

function viewAudit() {
  const card = el('div', { className: 'card' }, [el('h2', { text: 'Computer check' })])
  if (!state.audit) {
    card.appendChild(el('p', { text: 'Click “Check this PC” to scan.' }))
    return card
  }
  const ul = el('ul', { className: 'check-list' })
  for (const c of state.audit.checks || []) {
    ul.appendChild(
      el('li', {}, [
        el('div', { className: 'dot ' + (c.ok ? 'ok' : c.required ? 'bad' : 'warn') }),
        el('div', {}, [
          el('div', { className: 'check-title', text: c.label }),
          el('div', { className: 'check-detail', text: c.detail }),
          el('div', { className: 'check-detail', text: c.action })
        ])
      ])
    )
  }
  card.appendChild(ul)
  card.appendChild(el('div', { className: 'summary', text: state.audit.summary || '' }))
  card.appendChild(
    el('p', {
      className: 'credit',
      text: 'Note: Install/update only replaces program files. Your projects in Documents\\Slate are never deleted.'
    })
  )
  if (state.error) card.appendChild(el('div', { className: 'err', text: state.error }))
  return card
}

function viewInstall() {
  const title =
    state.mode === 'update' ? 'Updating Slate' : state.mode === 'repair' ? 'Repairing Slate' : 'Installing Slate'
  const card = el('div', { className: 'card' }, [el('h2', { text: title })])
  card.appendChild(
    el('div', {
      className: 'okbox',
      text:
        'Safety: Only the app under LocalAppData\\Programs\\Slate is replaced.\n' +
        'Your projects folder is protected:\n' +
        (state.paths.projectsDir || '%USERPROFILE%\\Documents\\Slate')
    })
  )
  card.appendChild(
    el('label', { className: 'opt' }, [
      el('input', {
        type: 'checkbox',
        checked: state.desktop,
        onChange: (e) => {
          state.desktop = e.target.checked
        }
      }),
      'Also put a shortcut on my Desktop'
    ])
  )
  const pct = state.progress.percent || 0
  card.appendChild(
    el('div', { className: 'progress-wrap' }, [
      el('div', { className: 'bar' }, [el('i', { style: `width:${pct}%` })]),
      el('div', {
        className: 'prog-label',
        text: state.progress.step || (state.busy ? 'Working…' : 'Ready when you are')
      }),
      el('div', {
        className: 'prog-detail',
        text: state.progress.detail || 'This can take 10–20 minutes the first time.'
      })
    ])
  )
  if (state.error) card.appendChild(el('div', { className: 'err', text: state.error }))
  if (state.result && !state.busy) {
    card.appendChild(el('div', { className: 'okbox', text: state.result.summary || 'Done.' }))
  }
  return card
}

function viewFinish() {
  const card = el('div', { className: 'card' }, [el('h2', { text: 'Finished' })])
  if (state.result) {
    card.appendChild(el('div', { className: 'okbox', text: state.result.summary || 'Slate is ready.' }))
  }
  card.appendChild(
    el('p', {
      text: 'Projects protected at: ' + (state.paths.projectsDir || 'Documents\\Slate')
    })
  )
  const brain = state.brain
  if (brain) {
    card.appendChild(el('h2', { text: brain.message || 'AI brain (optional)' }))
    const ul = el('ul', {})
    for (const s of brain.nextSteps || []) ul.appendChild(el('li', { text: s }))
    card.appendChild(ul)
  }
  if (state.error) card.appendChild(el('div', { className: 'err', text: state.error }))
  return card
}

function viewFooter() {
  const left = el('div', { className: 'btn-row' })
  const right = el('div', { className: 'btn-row' })

  if (state.view === 'deep-audit') {
    left.appendChild(
      el('button', {
        className: 'ghost',
        text: 'Back to home',
        onClick: () => {
          state.view = 'home'
          state.error = null
          render()
        }
      })
    )
    if (state.deep?.updates?.updateAvailable) {
      right.appendChild(
        el('button', {
          className: 'primary',
          text: 'Install update',
          onClick: () => startWizard('update')
        })
      )
    }
    right.appendChild(
      el('button', {
        text: 'Re-run audit',
        disabled: state.busy,
        onClick: () => runDeepAudit()
      })
    )
  } else if (state.view === 'home') {
    // actions on home card
  } else if (state.step === 1) {
    left.appendChild(
      el('button', {
        className: 'ghost',
        text: 'Home',
        onClick: () => {
          state.view = 'home'
          state.step = 0
          render()
        }
      })
    )
    right.appendChild(
      el('button', { text: 'Check again', disabled: state.busy, onClick: () => doAudit() })
    )
    right.appendChild(
      el('button', {
        className: 'primary',
        text: 'Continue',
        disabled: state.busy || !state.audit?.canProceed,
        onClick: () => {
          state.step = 2
          state.view = 'wizard'
          state.error = null
          render()
        }
      })
    )
  } else if (state.step === 2) {
    left.appendChild(
      el('button', {
        className: 'ghost',
        text: 'Back',
        disabled: state.busy,
        onClick: () => {
          state.step = 1
          render()
        }
      })
    )
    right.appendChild(
      el('button', {
        className: 'primary',
        text: state.busy
          ? 'Please wait…'
          : state.result
            ? 'Continue'
            : state.mode === 'update'
              ? 'Install update'
              : 'Install Slate',
        disabled: state.busy,
        onClick: async () => {
          if (state.result) {
            state.step = 3
            render()
            return
          }
          await doInstall()
        }
      })
    )
  } else if (state.step === 3) {
    left.appendChild(
      el('button', {
        className: 'ghost',
        text: 'Home',
        onClick: async () => {
          state.view = 'home'
          state.step = 0
          state.result = null
          try {
            state.installStatus = await go().GetInstallStatus()
          } catch (_) {}
          render()
        }
      })
    )
    right.appendChild(
      el('button', {
        className: 'primary',
        text: 'Open Slate',
        onClick: () => go()?.LaunchSlateProcess()
      })
    )
  }

  return el('div', { className: 'footer' }, [left, right])
}

async function startWizard(mode) {
  state.mode = mode || 'install'
  state.view = 'wizard'
  state.step = 1
  state.result = null
  state.error = null
  state.progress = { step: '', detail: '', percent: 0 }
  await doAudit()
}

async function doAudit() {
  state.busy = true
  state.error = null
  render()
  try {
    state.paths = (await go().GetPaths()) || state.paths
    state.audit = await go().RunAudit()
    state.installStatus = await go().GetInstallStatus()
  } catch (e) {
    state.error = String(e?.message || e)
  } finally {
    state.busy = false
    render()
  }
}

async function runDeepAudit() {
  state.view = 'deep-audit'
  state.busy = true
  state.error = null
  state.deep = null
  render()
  try {
    state.paths = (await go().GetPaths()) || state.paths
    state.deep = await go().RunDeepAudit()
    state.installStatus = state.deep.install || (await go().GetInstallStatus())
    state.updates = state.deep.updates || null
  } catch (e) {
    state.error = String(e?.message || e)
  } finally {
    state.busy = false
    render()
  }
}

async function checkUpdates() {
  state.busy = true
  state.error = null
  render()
  try {
    state.updates = await go().CheckForUpdates()
    state.installStatus = await go().GetInstallStatus()
  } catch (e) {
    state.error = String(e?.message || e)
  } finally {
    state.busy = false
    render()
  }
}

async function doInstall() {
  state.busy = true
  state.error = null
  state.result = null
  state.progress = { step: 'Starting', detail: 'Preparing (projects protected)…', percent: 1 }
  render()
  try {
    const fn = state.mode === 'update' ? go().StartUpdate : go().StartInstall
    const out = await fn(!!state.desktop)
    state.result = out.result
    state.brain = out.brain
    state.progress = { step: 'Done', detail: 'Finished. Projects untouched.', percent: 100 }
    state.step = 3
    state.installStatus = await go().GetInstallStatus()
  } catch (e) {
    state.error =
      String(e?.message || e) +
      '\n\nYour project files were not modified.\nLog: ' +
      (state.paths.logPath || '%TEMP%\\slate-install.log') +
      '\nYou can safely run Setup again.'
  } finally {
    state.busy = false
    render()
  }
}

eventsOn('install:progress', (payload) => {
  const p = payload || {}
  state.progress = {
    step: p.step || '',
    detail: p.detail || '',
    percent: p.percent || 0
  }
  if (state.step === 2) render()
})

bootstrap()
