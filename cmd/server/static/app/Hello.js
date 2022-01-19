define(function() {
  return {
    name: 'Hello',
    props: ['salutation'],
    emits: ['salute'],
    methods: {
      salute() {
        this.$emit('salute', this.salutation)
      },
    },
    template: `
      <a href="#" @click.prevent="salute">{{salutation}}, World!</a>
    `
  }
})
