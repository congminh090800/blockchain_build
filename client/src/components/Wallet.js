import { Box, Button, Container, TextField, Typography } from "@mui/material";
import React, { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { ArrowLeft } from "@mui/icons-material";
import axios from "axios";

const Wallet = () => {
  const { address } = useParams();
  const [data, setData] = useState({});
  const navigate = useNavigate();
  const [value, setValue] = useState("");
  const [sendA, setSendA] = useState("");

  React.useEffect(() => {
    (async () => {
      try {
        const res = await axios.get("http://localhost:8080/getbalance", {
          params: {
            address,
          },
        });
        setData(res.data);
      } catch (error) {
        navigate("/");
      }
    })();
  }, [address]);

  return (
    <Container
      fullWidth
      maxWidth="md"
      sx={{
        height: "100vh",
      }}
    >
      <Box className="df">
        <Button
          startIcon={<ArrowLeft />}
          onClick={() => {
            navigate("/");
          }}
        >
          GO BACK
        </Button>
      </Box>
      <Box className="df fdc">
        <Typography
          variant="h6"
          sx={{
            ml: 4,
          }}
          className="sb"
        >
          Wallet : {address}
        </Typography>
        <Typography
          variant="h6"
          sx={{
            ml: 4,
          }}
          className="sb"
        >
          Balance : {data.balance}
        </Typography>
        <Box className="df fdc" sx={{ p: 3 }}>
          <Typography>Send crypto to</Typography>

          <TextField
            label="address"
            value={sendA}
            onChange={(e) => {
              setSendA(e.target.value);
            }}
            sx={{ mt: 2 }}
          ></TextField>

          <TextField
            label="amount"
            type="number"
            value={value}
            onChange={(e) => {
              setValue(e.target.value);
            }}
            sx={{ mt: 2 }}
          ></TextField>
          <Box className="df" sx={{ justifyContent: "flex-end" }}>
            <Button
              sx={{ mt: 2, minWidth: 200 }}
              disabled={!data.balance}
              variant="contained"
              color="primary"
              onClick={async () => {
                try {
                  const res = await axios.post("http://localhost:8080/send", {
                    from: address,
                    to: sendA,
                    amount: value,
                  });
                  console.log(
                    "🚀 ~ file: Wallet.js ~ line 99 ~ onClick={ ~ res",
                    res
                  );
                  setData((prev) => ({
                    ...prev,
                    balance: prev.balance - value,
                  }));
                } catch (e) {}
              }}
            >
              Send
            </Button>
          </Box>
        </Box>
      </Box>
    </Container>
  );
};

export default Wallet;
