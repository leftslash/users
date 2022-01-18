define(function(require) {
  const Hello = require('app/Hello')
  return {
    name: 'App',
    components: {
      Hello
    },
    data() {
      return {
        message: null,
        salutations: [ 'Hello', 'Hola', 'Bonjour', 'Gutentag' ]
      }
    },
    methods: {
      salute(message) {
        this.message = message
      }
    },
    template: `
      <ul class="mt-3">
        <li v-for="salutation in salutations">
          <Hello @salute="salute" :salutation="salutation"/>
        </li>
      </ul>
      <div class="ms-3">
        {{message}}
      </div>
    `
  }
})
