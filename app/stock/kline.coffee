require 'main.css'
d3 = require 'd3'
defaults =
  container: 'body'
  margin:
    top: 20
    right: 50
    bottom: 30
    left: 50

extend = ->
  dest = {}
  for i in arguments when i
    for k,v of i
      dest[k] = dest[k] || v
  dest

parseDate = d3.time.format("%Y-%m-%dT%XZ").parse
formatValue = d3.format(",.2f")
fmtCent = (d) -> formatValue d/100

Plugins = {}

[cup, cdown, ceq] = ["#f00", "#080", "#000"]

kColor = (d, i, data) ->
  if d.open == d.close
    if i and data
      if data[i] and data[i-1]
        return cup if data[i].open >= data[i-1].close
        return cdown if data[i].open < data[i-1].close
    return ceq
  if d.open > d.close
    return cdown
  cup

class KLine
  constructor: (@options) ->
    @dispatch = d3.dispatch('resize', 'param', 'tip')
    @options = extend {}, @options, defaults
    @_data = []
    @_ui = {}
    @_left = 0
    @_max_left = 0
    @plugins = []
    @_param = {}
    @options.size = +@options.size || 100

  update_size: (size, left) ->
    @options.size = size || @options.size
    left = left || @_left
    atrightedge = @_left == @_max_left
    @_max_left = Math.max(0, @_data.length - 1 - @options.size)

    if atrightedge and left is @_left
      @_left = @_max_left
    else
      @_left = Math.min(@_max_left, Math.max(0, left))

  data: (data) ->
    if not arguments.length
      return @_data.slice(@_left, @_left+@options.size+1)

    s = data.id
    k = @param 'k'
    dataset = switch k
      when '1' then data.m1s
      when '5' then data.m5s
      when '30' then data.m30s
      when 'week' then data.weeks
      when 'month' then data.months
      else data.days
    @_dataset = data
    @_datasel = dataset
    data = dataset.data

    data.forEach (d) ->
      d.date = parseDate(d.time)
    @_data = data
    @update_size()

  param: (p) ->
    switch typeof p
      when 'object'
        for k,v of p
          @_param[k] = v
        @dispatch.param()
      when 'string'
        return @_param[p]
      else
        return @_param

  init: ->
    @stop()
    @initUI()
    @initPlugins()

    @on_event 'kdata', (data) =>
      @data data
      @draw()
    @dispatch.on 'param.core', =>
      d3.timer =>
        @data @_dataset
        @draw()

  add_plugin_obj: (plugin) ->
    @plugins.push plugin

  initPlugins: ->
    @plugins = []
    for n,c of Plugins
      plugin = new c @
      continue if plugin.init() is off
      @add_plugin_obj plugin

  resize: (w, h) ->
    w = w || @_ui.container[0][0].clientWidth
    h = h || window.screen.availHeight * 0.618
    @options.width = w - @options.margin.left - @options.margin.right
    @options.height = h - @options.margin.top - @options.margin.bottom
    @dispatch.resize()

  initUI: ->
    container = @_ui.container = d3.select @options.container || 'body'
    container.html('')
    width = parseInt container.style('width')
    if width < 1
      width = window.screen.availWidth
    height = parseInt container.style('height')
    if height < 1
      height = window.screen.availHeight * 0.618
    @options.width = width - @options.margin.left - @options.margin.right
    @options.height = height - @options.margin.top - @options.margin.bottom
    width = @options.width
    height = @options.height
    margin = @options.margin

    svg = @_ui.svg = container.append("svg")
      .attr("width", width + margin.left + margin.right)
      .attr("height", height + margin.top + margin.bottom)
      .append("g")
      .attr("transform", "translate(#{margin.left},#{margin.top})")

    x = @_ui.x = d3.scale.linear()
      .range([0, width])

    y = @_ui.y = d3.scale.linear()
      .range([height, 0])

    xAxisTickFormat = (i) =>
      data = @data()
      if typeof data[i] == 'undefined'
        return 'F'
      @prevTick = @prevTick || 0
      prevTick = @prevTick
      @prevTick = i

      if i == 0
        return d3.time.format("%Y-%m-%d")(data[i].date)
      if data[i].date == data[prevTick].date
        return ''
      if data[i].date.getYear() isnt data[prevTick].date.getYear()
        return d3.time.format("%Y-%m-%d")(data[i].date)
      if data[i].date.getMonth() isnt data[prevTick].date.getMonth()
        return d3.time.format("%m-%d")(data[i].date)
      if data[i].date.getDay() isnt data[prevTick].date.getDay()
        return d3.time.format("%d %H:%M")(data[i].date)
      if data[i].date.getHours() isnt data[prevTick].date.getHours()
        return d3.time.format("%H:%M")(data[i].date)
      d3.time.format(":%M")(data[i].date)

    xAxis = @_ui.xAxis = d3.svg.axis()
      .scale(x)
      .orient("bottom")
      .tickSize(-height, 0)
      .tickFormat(xAxisTickFormat)

    yAxis = @_ui.yAxis = d3.svg.axis()
      .scale(y)
      .orient("left")
      .ticks(6)
      .tickSize(-width)
      .tickFormat(fmtCent)

    svg.append("g")
      .attr("class", "x axis")
      .attr("transform", "translate(0, #{height})")
      .call(xAxis)

    svg.append("g")
      .attr("class", "y axis")
      .call(yAxis)

    zoomed = =>
      n = zoom.scale()
      @zs = @zs || n
      o = @zs
      @zs = n

      x1 = zoom.translate()[0]
      @zx = @zx || 0
      x0 = @zx
      @zx = x1

      nsize = @options.size
      nleft = @_left
      if n < o
        nsize = parseInt nsize * 1.1
      else if n > o
        nsize = parseInt nsize * 0.9
      else
        if x0 > x1
          nleft = nleft + Math.max(20, parseInt nsize * 0.1)
        else if x0 < x1
          nleft = nleft - Math.max(20, parseInt nsize * 0.1)
        else
          return
      @update_size(nsize, nleft)

      fn = => @draw()
      d3.timer fn

    zoom = d3.behavior.zoom()
      .on("zoom", zoomed)

    svg
      .call(zoom)
    svg.append("rect")
      .attr("class", "pane")
      .attr("width", width)
      .attr("height", height)

    @dispatch.on 'resize.core', =>
      width = @options.width
      height = @options.height
      margin = @options.margin
      container.select('svg')
        .attr("width", width + margin.left + margin.right)
        .attr("height", height + margin.top + margin.bottom)
      x.range([0, width])
      y.range([height, 0])
      xAxis.tickSize(-height, 0)
      yAxis.tickSize(-width)
      svg.select('.x.axis').attr("transform", "translate(0, #{height})")
      svg.select('rect.pane')
        .attr("width", width)
        .attr("height", height)
      fn = => @draw()
      d3.timer fn

  draw: ->
    x = @_ui.x
    y = @_ui.y
    data = @data()

    x.domain([0, data.length-1])
    y.domain([d3.min(data, (d)->d.low) * 0.99, d3.max(data, (d)->d.high)])

    @update data, @_datasel, @_dataset

  updateAxis: ->
    @_ui.svg.select(".x.axis").call(@_ui.xAxis)
    @_ui.svg.select(".y.axis").call(@_ui.yAxis)

  update: (data, datasel, dataset) ->
    @updateAxis()
    for plugin in @plugins when plugin.updateAxis
      plugin.updateAxis data, datasel, dataset
    for plugin in @plugins
      plugin.update data, datasel, dataset

  init_websocket: (done) ->
    if @io
      @io.on('ready', done)
      if @io.connected
        return done()
      return

    io = {}
    handles = {}
    @io = io

    ev =
      s: @param 's'
      k: @param 'k'
      fq: @param 'fq'

    connect = ->
      protocol = if location.protocol.toLowerCase() == 'https:' then 'wss:' else 'ws:'
      ws = new WebSocket("#{protocol}//#{location.host}/socket.io/")
      io.ws = ws
      ws.onopen = (evt) ->
        io.connected = true
        io.trigger('ready', evt)
        io.ws.send(JSON.stringify(ev))

      ws.onclose = (evt) ->
        io.connected = false
        io.trigger('close', evt)

      ws.onmessage = (evt) ->
        res = JSON.parse(evt.data)
        io.trigger('data', res)

      ws.onerror = (evt) ->
        io.trigger('error', evt)

    io.on = (event, cb) ->
      handles[event] = handles[event] || []
      i = handles[event].indexOf(cb)
      if i < 0
        handles[event].push(cb)

    io.off = (event, cb) ->
      return unless handles[event]
      i = handles[event].indexOf(cb)
      if i > -1
        handles[event].splice i, 1

    io.emit = (event, data) ->
      msg = event: event, data: data
      io.ws.send(JSON.stringify(msg))

    io.trigger = (event, data) ->
      return unless handles[event]
      fn(data) for fn in handles[event]

    io.on 'data', (data) ->
      for event in ['kdata']
        ename = [ev.s,ev.k,ev.fq,event].join('.')
        data.param = ev
        io.trigger ename, data

    if done
      @io.on 'ready', done
    onclose = (evt) ->
      console.log evt
      setTimeout((->connect()), 1000)

    io.on 'close', onclose

    io.close = ->
      io.off 'close', onclose
      io.ws.close()

    io.connect = connect

  on_event: (event, cb) ->
    @init_websocket =>
      s = @param 's'
      k = @param 'k'
      fq = @param 'fq'
      ename = [s,k,fq,event].join('.')
      @io.on(ename, cb)

  start: ->
    @io.connect()
  stop: ->
    if @io
      @io.close()
      @io = off

  notification: (id, msg) ->
    unless window.Notification
      console.log('bro, your browser is not support notification')
      return
    Notification.requestPermission (permission) ->
      config =
        body: msg
        dir:'auto'
      notification = new Notification(id, config)

KLine.register_plugin = (name, clazz) ->
  Plugins[name] = clazz

KLine.extend = extend
KLine.kColor = kColor

filter = (src, range) ->
  if (src||[]).length < 1
    return []
  if (range||[]).length < 2
    return []

  start = range[0].date
  end = range[range.length-1].date

  hash = {}
  hash[+d.date] = i for d, i in range

  closeIndexOf = (date) ->
    date = +date
    j = 0
    j = i for d, i in range when date >= +d.date
    j

  indexOf = (date) ->
    date = +date
    if +range[0].date > date
      return -1
    if +range[range.length-1].date < date
      return -1
    hash[+date] || closeIndexOf(date)

  for d in src
    d.date = d.date || parseDate(d.Time)
    d.edate = d.edate || parseDate(d.ETime)
    d.i = indexOf d.date
    d.ei = indexOf d.edate

  rv = []
  for d in src
    if d.i != -1
      rv.push d
      continue

    if rv.length < 2
      rv = [d]
      continue

    if rv[rv.length-1].i != -1
      rv.push d
      continue
  rv
KLine.filter = filter

module.exports = KLine