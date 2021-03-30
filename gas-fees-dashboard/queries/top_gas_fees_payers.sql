-- Top 100 Gas Fees Payers
-- Top 100 Gas Fees Payers
-- {"type":"TABLE","name":"Top Gas Fees Payers Table","description":"","options":{"itemsPerPage":10,"columns":[{"numberFormat":"0,0.00","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"fees_paid","type":"float","displayAs":"number","visible":true,"order":100000,"title":"Fees Paid","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"from","type":"string","displayAs":"string","visible":true,"order":100001,"title":"Address","allowSearch":true,"alignContent":"left","allowHTML":true,"highlightLinks":false}]},"query_id":29}
--
SELECT sum(least((gas_premium * gas_limit) + (base_fee * gas_used), gas_limit * gas_fee_cap)) * 0.000000000000000001 AS fees_paid,
       from_id
FROM filecoin.messages AS msg
WHERE gas_used > 0
GROUP BY from_id
ORDER BY fees_paid DESC
LIMIT 100;