Vue = require 'vue'
KLine = require '../stock/webpack'
config = require './config'

Vue.directive 'kanpan',
  deep: true
  bind: ->
    window.addEventListener 'resize', =>
      @kl.resize() if @kl
    window.addEventListener 'keyup', (e) =>
      return if e.target.tagName == 'INPUT'
      return unless @kl
      handles =
        49: 'nmas'
        50: 'nc'
        51: 'nvolume'
        52: 'nmacd'
        72: 'handcraft'
      name = handles[e.keyCode] || off
      if name
        param = {}
        param[name] = not @kl.param name
        config.update param
        @kl.param param
    window.addEventListener 'keydown', (e) =>
      return if e.target.tagName == 'INPUT'
      return unless @kl
      kl = @kl
      move_handles =
        35: -> kl.move_to('end')
        36: -> kl.move_to('home')
        37: -> kl.move_to('left')
        39: -> kl.move_to('right')
      ctl = move_handles[e.keyCode] || ->
      ctl()

  update: (value, oldValue) ->
    return unless value
    return unless value.s
    settings = config.load() || {}
    params = JSON.parse(JSON.stringify(value))
    for k,v of params
      settings[k] = v

    if @kl
      console.log 'has kl'
      if value.s is @kl.param 's'
        return @kl.param settings
      @kl.stop()
    else
      console.log 'new kl'
      @kl = new KLine(container: @el)
    kl = @kl
    kl.param settings

    setTimeout ->
      kl.init()
      kl.start()
    , 500