-- Gas Fees For Epoch
-- Sums of Gas Fees and premiums paid to miners for every message in each epoch of last week
-- {"type":"CHART","name":"Chart","description":"","options":{"globalSeriesType":"line","sortX":true,"legend":{"enabled":true},"yAxis":[{"type":"linear","title":{"text":"FIL"},"rangeMax":0.002},{"type":"linear","opposite":true}],"xAxis":{"type":"-","labels":{"enabled":true},"title":{"text":"epoch"}},"error_y":{"type":"data","visible":true},"series":{"stacking":null,"error_y":{"type":"data","visible":true},"percentValues":false},"seriesOptions":{"premiums":{"zIndex":1,"index":0,"type":"line","yAxis":0},"gas_fees":{"zIndex":0,"index":0,"type":"line","yAxis":0}},"valuesOptions":{},"columnMapping":{"premiums":"y","epoch":"x","gas_fees":"y"},"direction":{"type":"counterclockwise"},"numberFormat":"0,0[.]00000","percentFormat":"0[.]00%","textFormat":"","missingValuesAsZero":true,"showDataLabels":false,"dateTimeFormat":"DD/MM/YY HH:mm","customCode":"// Available variables are x, ys, element, and Plotly\n// Type console.log(x, ys); for more info about x and ys\n// To plot your graph call Plotly.plot(element, ...)\n// Plotly examples and docs: https://plot.ly/javascript/"},"query_id":9}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":3,"sizeY":8,"minSizeX":1,"maxSizeX":6,"minSizeY":5,"maxSizeY":1000,"col":0,"row":22}},"text":"","width":1,"dashboard_id":1,"visualization_id":2}
SELECT sum(least((gas_premium * gas_limit) + (base_fee * gas_used), gas_limit * gas_fee_cap)) * 0.000000000000000001 AS gas_fees,
       sum(least((gas_premium * gas_limit) + (base_fee * gas_used), gas_limit * gas_fee_cap) - (base_fee * gas_used)) * 0.000000000000000001 AS premiums,
       height AS epoch
FROM filecoin.messages
WHERE gas_used > 0
  AND height >
    (SELECT max(height)
     FROM filecoin.tipsets) - 20160
GROUP BY epoch;