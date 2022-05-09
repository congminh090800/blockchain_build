import { ArrowLeft } from "@mui/icons-material";
import { Box, Button, Container, Typography } from "@mui/material";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";

const History = () => {
  const navigate = useNavigate();
  const [data, setData] = useState([]);
  console.log("🚀 ~ file: History.js ~ line 10 ~ History ~ data", data);
  React.useEffect(() => {
    (async () => {
      const res = await axios.get("http://localhost:8080/print");
      setData(res.data.data);
    })();
  }, []);

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
      <Typography variant="h6" className="sb" sx={{ p: 2 }}>
        Block history
      </Typography>
      <Box className="df fdc">
        {data.map((item, index) => {
          return (
            <Box
              key={index}
              sx={{
                p: 2,
                m: "16px 0px",
                borderRadius: 2,
                border: "1px solid #e8e8e8",
              }}
            >
              <Typography>Hash : {item.Hash}</Typography>
              <Typography>Timestamp : {item.Timestamp}</Typography>
              <Typography>Prevhash: {item.PreviousHash}</Typography>
            </Box>
          );
        })}
      </Box>
    </Container>
  );
};

export default History;
