import d3 from 'd3';
import util from './util';
import KLine from './kline';

const defaults = {
  tmpl: '开盘价：<%- p.open%><br/>收盘价：<%- p.close%><br/>最高价：<%- p.High%><br/>最低价：<%- p.Low%>',
  margin: [0, 10, 0, 10],
  style: {
    padding: '6px 10px 4px',
    'line-height': '20px',
    position: 'absolute',
    'font-size': '12px',
    color: '#fff',
    background: 'rgb(57, 157, 179)',
    'border-radius': '10px',
    'box-shadow': '0px 0px 3px 3px rgba(0, 0, 0, 0.3)',
    opacity: '0.8',
    display: 'none',
  },
  width: 120,
  height: 82,
  x: {
    style: {
      display: 'inline-block',
      position: 'absolute',
      'font-size': '12px',
      background: '#FFF',
      'text-align': 'center',
      border: '1px solid #E3F4FF',
    },
    width: 95,
    format: 'YYYY/MM/DD HH:mm',
  },
  y: {
    style: {
      display: 'inline-block',
      position: 'absolute',
      'font-size': '12px',
      background: '#FFF',
      border: '1px solid #E3F4FF',
    },
  },
};

const formatValue = d3.format(',.2f');

function fmtCent(d) {
  return formatValue(d / 100);
}

function templ(name, d) {
  const e = d3.select(this);
  switch (name) {
    case 'k':
      return `${d.time}<br/>
        open: ${fmtCent(d.open)}<br/>
        high: ${fmtCent(d.High)}<br/>
        low: ${fmtCent(d.Low)}<br/>
        close: ${fmtCent(d.close)}<br/>
        volume: ${d.volume}`;
    case 'typing':
    case 'segment':
      return `${d.Time}<br/>
        high: ${fmtCent(d.High)}<br/>
        low: ${fmtCent(d.Low)}<br/>
        ${name}`;
    case 'hub':
      return `${d.Time} -- ${d.ETime}<br/>
        high: ${fmtCent(d.High)}<br/>
        low: ${fmtCent(d.Low)}<br/>
        ${name}`;
    case 'segment_line':
    case 'typing_line':
      return `${d.Time} -- ${d.ETime}<br/>
        high: ${fmtCent(d.High)}<br/>
        low: ${fmtCent(d.Low)}<br/>
        MACD: ${d.MACD}<br/>
        ${name}`;
    case 'mas':
      return `${e.attr('id')}
      <div style='background-color:${e.style('stroke')}'>${e.attr('id')}</div>
        ${name}`;
    default:
      return 'no templ';
  }
}

class KLineToolTip {
  constructor(root) {
    this.root = root;
    this.options = KLine.extend({}, this.root.options.tooltip, defaults);
  }

  init() {
    const tips = this.tips = this.root._ui.container.append('div')
      .attr('id', 'tooltip');

    const ref = this.options.style;
    for (const k in ref) {
      if (ref.hasOwnProperty(k)) {
        tips.style(k, ref[k]);
      }
    }

    const left = () => {
      const w = util.w();
      const tw = tips[0][0].clientWidth;
      let v = d3.event.pageX;
      if (w - tw - d3.event.pageX - 30 < 0) {
        v = d3.event.pageX - tw - 10;
      }
      return `${v}px`;
    };

    const top = () => {
      const h = util.h();
      const th = tips[0][0].clientHeight;
      let v = d3.event.pageY + 30;
      if (h - th - d3.event.pageY - 30 < 0) {
        v = d3.event.pageY - th - 30;
      }
      return `${v}px`;
    };

    this.root.dispatch.on('tip', (e, ...args) => {
      tips
        .style('display', '')
        .style('left', left)
        .style('top', top)
        .html(templ.apply(e, args))
        .transition()
        .duration(5000)
        .transition()
        .style('display', 'none');
    });
  }

  update() {
  }
}

KLine.register_plugin('tooltip', KLineToolTip);
