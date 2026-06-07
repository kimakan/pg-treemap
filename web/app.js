const chart = echarts.init(document.getElementById('main'), 'vintage');

function formatBytes(bytes) {
  if (!bytes) return '0 B';

  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));

  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

function getLevelOption() {
  return [
    {
      itemStyle: {
        borderColor: '#bbb',
        borderWidth: 0,
        gapWidth: 2
      },
      upperLabel: {
        show: false
      }
    },
    {
      itemStyle: {
        borderColor: '#aaa',
        borderWidth: 5,
        gapWidth: 2
      },
      emphasis: {
        itemStyle: {
          borderColor: '#ddd'
        }
      }
    },
    {
      colorSaturation: [0.35, 0.5],
      itemStyle: {
        borderWidth: 5,
        gapWidth: 1,
        borderColorSaturation: 0.8
      }
    }
  ];
}

fetch('/api')
  .then(r => r.json())
  .then(data => {
    chart.setOption({
      title: {
        text: 'Daiquiri Databases',
        left: 'center'
      },
      tooltip: {
        formatter: function (info) {
          var value = info.value;
          var treePathInfo = info.treePathInfo;
          var treePath = [];
          for (var i = 1; i < treePathInfo.length; i++) {
            treePath.push(treePathInfo[i].name);
          }
          return [
            '<div class="tooltip-title">' +
            echarts.format.encodeHTML(treePath.join('.')) +
            '</div>',
            'Total Size:' + echarts.format.addCommas(formatBytes(value))
          ].join('');
        }
      },
      series: [{
        label: {
          show: true,
          fontSize: 18,
          formatter: function (params) {
            return `${params.name} (${formatBytes(params.data.value)})`;
          }
        },
        upperLabel: {
          show: true,
          fontWeight: 'bold',
          fontSize: 22,
          height: 40
        },
        itemStyle: {
          borderColor: '#eee'
        },
        levels: getLevelOption(),
        type: 'treemap',
        leafDepth: 2,
        data: data
      }]
    });
  });
