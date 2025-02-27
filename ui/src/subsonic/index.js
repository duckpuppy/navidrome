const subsonicUrl = (command, id, options) => {
  const params = new URLSearchParams()
  params.append('u', localStorage.getItem('username'))
  params.append('t', localStorage.getItem('subsonic-token'))
  params.append('s', localStorage.getItem('subsonic-salt'))
  params.append('f', 'json')
  params.append('v', '1.8.0')
  params.append('c', 'NavidromeUI')
  params.append('id', id)
  if (options) {
    if (options.ts) {
      options['_'] = new Date().getTime()
      delete options.ts
    }
    Object.keys(options).forEach((k) => {
      params.append(k, options[k])
    })
  }
  return `rest/${command}?${params.toString()}`
}

export { subsonicUrl }
