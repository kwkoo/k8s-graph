const radius = 20
const repellingStrength = -50
const linkLength = 100
const maxLabelLength = 25

var color = d3.scaleOrdinal(d3.schemeCategory10)

var app = new Vue({
    el: '#app',

    data: {
        screen: 'loading',
        main: {
            graph: { "nodes": [], "links": [] },
            svg: {},
            width: 0,
            height: 0,
            g: '',
            simulation: {},
            linkElements: [],
            nodeElements: [],
            textElements: []
        },
    },

    mounted: function() {
        // todo: initiate call to load projects
        this.initGraph()
    },

    methods: {
        initGraph: function() {
            let that = this

            this.main.svg = d3.select("svg")
                .call(d3.zoom().on("zoom", function () {
                    that.main.svg.attr("transform", d3.event.transform)
                }))

            this.main.width = +this.main.svg.attr("width"),
            this.main.height = +this.main.svg.attr("height")

            this.getGraphData('dummy')
        },

        getGraphData: function(namespace) { // todo: currently unused
            let that = this

            d3.json("/api/graph", function(error, data) {
                // todo: check for error here
                if (error) throw error;

                // todo: check for data.error
            
                that.main.graph = data;

                that.main.simulation = d3.forceSimulation()
                    .force("link", d3.forceLink().distance(linkLength).id(function (d) { // distance set length of links
                        return d.id
                    }))
                    .force("charge", d3.forceManyBody().strength(repellingStrength)) // strength sets repelling force
                    .force("center", d3.forceCenter(that.main.width / 2, that.main.height / 2))
                
                that.main.g = that.main.svg.append("g").attr("class", "everything")

                that.main.linkElements = that.main.svg.append("g").selectAll(".link")
                that.main.nodeElements = that.main.svg.append("g").selectAll(".nodes")
                that.main.textElements = that.main.svg.append("g").selectAll(".texts")

                that.drawGraph()

                that.main.simulation.nodes(that.main.graph.nodes).on("tick", that.ticked)
                that.main.simulation.force("link").links(that.main.graph.links)
            })
        },


        drawGraph: function() {
            let that = this

            // empty current Graph contents
            this.main.g.html('')

            this.main.linkElements = this.main.g.append("g")
                .attr("class", "links")
                .selectAll("line")
                .data(this.main.graph.links)
                .enter().append("line")
                .style("stroke-width", 3)
                .style("stroke", "grey")
        
        
            this.main.nodeElements = this.main.g.append("g")
                .attr("class", "nodes")
                .selectAll("circle")
                .data(this.main.graph.nodes)
                .enter().append("circle")
                .attr("r", radius)  // adjust this value to set radius of node
                .attr("stroke", "#fff")
                .attr('stroke-width', 21)
                .attr("id", function (d) {
                    return d.id
                })
                .attr("fill", function (d) {
                    return color(d.kind)
                })
                .on("click", this.selectNode)
                .call(d3.drag()
                    .on("start", this.dragstarted)
                    .on("drag", this.dragged)
                    .on("end", this.dragended))
        
            this.main.textElements = this.main.g.append("g")
                .attr("class", "texts")
                .selectAll("text")
                .data(this.main.graph.nodes)
                .enter().append("text")
                .attr("text-anchor", "end")
                .text(function (node) {
                    let label = node.kind + "/" + node.name;
                    if (label.length > maxLabelLength) label = label.substring(0, maxLabelLength - 3) + "...";
                    return label;
                })
                .attr("font-size", 55)
                .attr("font-family", "sans-serif")
                .attr("fill", "black")
                .attr("style", "font-weight:bold;")
                .attr("dy", 40) // vertical position of text
                .attr("text-anchor", "middle")
        },

        ticked: function() {
            let that = this

            this.main.linkElements
                .attr("x1", function (d) {
                    return d.source.x
                })
                .attr("y1", function (d) {
                    return d.source.y
                })
                .attr("x2", function (d) {
                    return d.target.x
                })
                .attr("y2", function (d) {
                    return d.target.y
                })
            this.main.nodeElements
                .attr("cx", function (d) {
                    return d.x = Math.max(radius, Math.min(that.main.width - radius, d.x))
                })
                .attr("cy", function (d) {
                    return d.y = Math.max(radius, Math.min(that.main.height - radius, d.y))
                })
                .each(d => {
                    d3.select('#t_' + d.id).attr('x', d.x + 10).attr('y', d.y + 3)
                })
            this.main.textElements
                .attr('x', function (d) {
                    return d.x
                })
                .attr('y', function (d) {
                    return d.y
                })
        },

        selectNode: function(d) {
            console.log(d)
        },

        dragstarted: function(d) {
            if (!d3.event.active) this.main.simulation.alphaTarget(0.3).restart()
            d.fx = d.x
            d.fy = d.y
        },

        dragged: function(d) {
            d.fx = d3.event.x
            d.fy = d3.event.y
        },

        dragended: function(d) {
            if (!d3.event.active) this.main.simulation.alphaTarget(0)
            d.fx = null
            d.fy = null
        },
    }
})
