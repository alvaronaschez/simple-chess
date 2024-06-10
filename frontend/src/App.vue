<script setup lang="ts">
import { ref } from 'vue'
import { BoardApi, TheChessboard, type MoveEvent } from 'vue3-chessboard'
import 'vue3-chessboard/style.css'

let board: BoardApi
const color = ref()

const socket = new WebSocket('ws://localhost:5555/ws')
socket.addEventListener('message', (event) => {
  const message = JSON.parse(event.data)
  if (message.type === 'start') {
    color.value = message.color
  } else if (message.type === 'move') {
    const { from, to, promotion } = message
    board.move({ from, to, promotion })
  }
})

function handleBoardCreated(boardApi: BoardApi) {
  board = boardApi
}

function handleMove(move: MoveEvent) {
  if (!color.value.startsWith(move.color)) {
    return
  }
  const { from, to, promotion } = move
  const message = JSON.stringify({ from, to, promotion, color: color.value, type: 'move' })
  socket.send(message)
}
</script>

<template>
  <TheChessboard
    v-if="color"
    @move="handleMove"
    @board-created="handleBoardCreated"
    :player-color="color"
    :board-config="{ orientation: color }"
  />
  <h1 v-else>Waiting for player 2</h1>
</template>
