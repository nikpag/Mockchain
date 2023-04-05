<html>

<?php
function sendrequest(string $request, string $type)
{
  $curl = curl_init();
  curl_setopt_array($curl, array(
    CURLOPT_URL => $request,
    CURLOPT_RETURNTRANSFER => true,
    CURLOPT_ENCODING => '',
    CURLOPT_MAXREDIRS => 10,
    CURLOPT_TIMEOUT => 0,
    CURLOPT_FOLLOWLOCATION => true,
    CURLOPT_HTTP_VERSION => CURL_HTTP_VERSION_1_1,
    CURLOPT_CUSTOMREQUEST => $type,
  ));

  $response = curl_exec($curl);
  $responseinfo = curl_getinfo($curl);
  $httpcode = $responseinfo['http_code'];
  curl_close($curl);

  $arr = array("response" => $response, "code" => $httpcode);
  return $arr;
}

?>

</html>