<!DOCTYPE html>
<html lang="en">


<?php include "func/function.php"; ?>

<?php include "func/head.php"; ?>


<body class="d-flex flex-column min-vh-100 text-light" style="background-color: #111111;">

  <?php include "func/navbar.php";

  $port = "83.212.80.139:58080/";
  $response = sendrequest('http://' . $port . 'balance', 'GET');

  if ($response["code"] == 200) {
    $balance = $response["response"];
  } else {
    $balance = 55;
  }

  ?>

  <br>
  <div class="container justify-content-space-between">
    <h4>Transactions</h4> <br>

    <div class="row" style="width: 100%;">

      <div class="col">
        <h5>Make a transaction</h5>
        <br>
        <p>Make a transaction as an admin. Choose processes and the amount of NBC to be transferred:</p>
        <br>
        <div>
          <form action="<?php echo htmlspecialchars($_SERVER["PHP_SELF"]); ?>" method="POST">
            <div class="mb-3" style="width: 55%;">
              <label for="receiver" class="form-label">Receiver</label>
              <input class="form-control" name="receiver" id="receiver">
            </div>
            <div class="mb-3" style="width: 55%;">
              <label for="amount" class="form-label">Amount</label>
              <input class="form-control" name="amount" id="amount">
            </div>
            <br>
            <div class="container d-flex justify-content-center">
              <button class="btn btn-light" type="submit">Send</button>
            </div>
          </form>
        </div>

      </div>
      <div class="col card d-flex " style="background-color: #111111;">
        <div class="card-header d-flex justify-content-left">
          <h5>Wallet balance</h5>
        </div>
        <div class="card-body d-flex justify-content-start text-white">
          <p><b style="font-size: 10vw;"><?php echo $balance ?></b>
            <b style="font-size: 5vw;">NBC</b>
            <br>
          </p>
        </div>
        <div class="overflow-auto" style="max-width: 100%; max-height: 200px; background-color: #222222">
          <?php
          $port = "83.212.80.139:58080/";
          $response = sendrequest('http://' . $port . 'history', 'GET');

          if ($response["code"] == 200) {
            $balance = $response["response"];

            $json = json_decode($response["response"]);

            echo '<table class="table table-hover caption-top">
      <caption>History:</caption>
      <thead class = "table-dark text-white">
      <tr>
      <th scope="col">#</th>
      <th scope="col">From/To</th>
      <th scope="col">Node</th>
      <th scope="col">Amount</th>
      </tr>
      </thead>
      <tbody>';
            $i = 0;
            foreach ($json->transactionList as $key => $transaction) {

              echo     '<tr class = "text-white">
        <th scope="row">' . $i . '</th>
        <td>' . $transaction->fromTo . '</td>
        <td>' . $transaction->node . '</td>
        <td>' . $transaction->amount . '</td>
        </tr>';
              $i += 1;
            }
            echo "</table>";
          }
          ?>

        </div>

      </div>
      <?php
      $port = "83.212.80.139:58080/";
      if ($_SERVER["REQUEST_METHOD"] == 'POST') {
        $flag = 0;
        if (empty($_POST["receiver"])) {
          $flag = 1;
        } else {
          $receiver = $_POST["receiver"];
        }
        if (empty($_POST["amount"])) {
          $flag = 1;
        } else {
          $amount = $_POST["amount"];
        }
        if ($flag == 1) {
      ?>
          <br><br>
          <div class="container  card d-flex justify-content-center" style="background-color: #222222;">
            <div class="card-header d-flex justify-content-center">
              Response
            </div>
            <div class="card-body d-flex justify-content-center bg-danger text-white">
              All fields required!
            </div>
          </div>
          <?php
        } else {

          $response = sendrequest('http://' . $port . 'transaction/id0/' . $receiver . '/' . $amount, 'POST');
          if ($response["code"] == 200) {
          ?>
            <br><br>
            <div class="container  card d-flex justify-content-center" style="background-color: #222222;">
              <div class="card-header d-flex justify-content-center">
                Response
              </div>
              <div class="card-body d-flex justify-content-center bg-success text-white">
                Success!
              </div>
            </div>
          <?php
          } else {
          ?>
            <br><br>
            <div class="container  card d-flex justify-content-center" style="background-color: #222222;">
              <div class="card-header d-flex justify-content-center">
                Response
              </div>
              <div class="card-body d-flex justify-content-center bg-success text-white">
                Success!
              </div>
            </div>
      <?php
          }
        }
      }
      ?>
    </div>
  </div>



</body>

<?php include "func/footer.php"; ?>

</html>