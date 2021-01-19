-- Blocks And Win Counts By Miner
--
-- {"type":"CHART","name":"Blocks And Win Counts By Miner","description":"","options":{"globalSeriesType":"column","sortX":true,"legend":{"enabled":true},"yAxis":[{"type":"linear"},{"type":"linear","opposite":true}],"xAxis":{"type":"-","labels":{"enabled":true},"title":{"text":"miner"}},"error_y":{"type":"data","visible":true},"series":{"stacking":null,"error_y":{"type":"data","visible":true}},"seriesOptions":{"blocks_cnt":{"type":"column","yAxis":0,"zIndex":0,"index":0},"win_counts":{"type":"column","yAxis":0,"zIndex":1,"index":0}},"valuesOptions":{},"columnMapping":{"miner":"x","blocks_cnt":"y","win_counts":"y"},"direction":{"type":"counterclockwise"},"numberFormat":"0,0[.]00000","percentFormat":"0[.]00%","textFormat":"","missingValuesAsZero":true,"showDataLabels":false,"dateTimeFormat":"DD/MM/YY HH:mm","customCode":"// Available variables are x, ys, element, and Plotly\n// Type console.log(x, ys); for more info about x and ys\n// To plot your graph call Plotly.plot(element, ...)\n// Plotly examples and docs: https://plot.ly/javascript/"},"query_id":28}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":6,"sizeY":8,"maxSizeY":1000,"maxSizeX":6,"minSizeY":5,"minSizeX":1,"col":0,"row":115}},"text":"","width":1,"dashboard_id":1,"visualization_id":56}
SELECT miner,
       count(*) AS blocks_cnt,
       sum(win_count) AS win_counts
FROM filecoin.blocks
GROUP BY miner;
