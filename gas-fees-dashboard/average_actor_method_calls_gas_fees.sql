-- Average Actor Method Calls Gas Fees
-- Table with averages of gas fees and it's component for every distinct actor's method call for last 24 hours
-- {"type":"TABLE","name":"Table","description":"","options":{"itemsPerPage":10,"columns":[{"numberFormat":"0,0","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"method","type":"integer","displayAs":"number","visible":true,"order":100000,"title":"method","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"to","type":"string","displayAs":"string","visible":true,"order":100001,"title":"to","allowSearch":true,"alignContent":"left","allowHTML":false,"highlightLinks":false},{"numberFormat":"0,0.00","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"avg_value","type":"float","displayAs":"number","visible":true,"order":100002,"title":"average value","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0.00","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"avg_gas_fees","type":"float","displayAs":"number","visible":true,"order":100003,"title":"average gas fees","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0.00","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"avg_gas_limit","type":"float","displayAs":"number","visible":true,"order":100004,"title":"average gas limit","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0.00","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"avg_gas_used","type":"float","displayAs":"number","visible":true,"order":100005,"title":"average gas used","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0.00","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"avg_gas_fee_cap","type":"float","displayAs":"number","visible":true,"order":100006,"title":"average gas fee cap","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0.00","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"avg_gas_premium","type":"float","displayAs":"number","visible":true,"order":100007,"title":"average gas premium","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0.00","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"avg_parent_base_fee","type":"float","displayAs":"number","visible":true,"order":100008,"title":"average parent base fee","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false}]},"query_id":5}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":true,"sizeX":6,"sizeY":11,"maxSizeY":1000,"maxSizeX":6,"minSizeY":1,"minSizeX":2,"col":0,"row":54}},"text":"","width":1,"dashboard_id":1,"visualization_id":9}
SELECT method,
       "to",
       avg(value) AS avg_value,
       avg(gas_fees) AS avg_gas_fees,
       avg(gas_limit) AS avg_gas_limit,
       avg(gas_used) AS avg_gas_used,
       avg(gas_fee_cap) AS avg_gas_fee_cap,
       avg(gas_premium) AS avg_gas_premium,
       avg(parent_base_fee) AS avg_parent_base_fee
FROM (
    SELECT method, "from"
           "to",
           least((gas_premium * gas_limit) + (parent_base_fee * gas_used), gas_limit * gas_fee_cap) AS gas_fees,
           value,
           gas_limit,
           gas_used,
           gas_fee_cap,
           gas_premium,
           parent_base_fee,
           block_cid,
           height AS epoch,
           msg.cid AS msg_cid
    FROM filecoin.messages AS msg
    LEFT JOIN filecoin.blocks blk ON blk.cid = block_cid
    WHERE gas_used > 0 AND method > 0
    ORDER BY epoch DESC
    limit 2880
) AS g
GROUP BY method, "to"