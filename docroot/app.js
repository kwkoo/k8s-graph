var radius = 20;
var repellingStrength = -50;
var linkLength = 100;
var maxLabelLength = 25;
var graph = { "nodes": [], "links": [] };

var svg = d3.select("svg")
    .call(d3.zoom().on("zoom", function () {
        svg.attr("transform", d3.event.transform)
    })),
    width = +svg.attr("width"),
    height = +svg.attr("height")

var g;
var simulation;
var linkElements, nodeElements, textElements;
var color = d3.scaleOrdinal(d3.schemeCategory10);

d3.json("/api/graph", function (error, data) {
    // todo: check for error here
    if (error) throw error;

    graph = data;

    simulation = d3.forceSimulation()
        .force("link", d3.forceLink().distance(linkLength).id(function (d) { // distance set length of links
            return d.id;
        }))
        .force("charge", d3.forceManyBody().strength(repellingStrength)) // strength sets repelling force
        .force("center", d3.forceCenter(width / 2, height / 2));

    g = svg.append("g")
        .attr("class", "everything");

    linkElements = svg.append("g").selectAll(".link");
    nodeElements = svg.append("g").selectAll(".nodes");
    textElements = svg.append("g").selectAll(".texts");

    drawGraph();


    simulation
        .nodes(graph.nodes)
        .on("tick", ticked);

    simulation.force("link")
        .links(graph.links);

});


function drawGraph() {

    // empty current Graph contents
    g.html('')


    linkElements = g.append("g")
        .attr("class", "links")
        .selectAll("line")
        .data(graph.links)
        .enter().append("line")
        .style("stroke-width", 3)
        .style("stroke", "grey")


    nodeElements = g.append("g")
        .attr("class", "nodes")
        .selectAll("circle")
        .data(graph.nodes)
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
        .on("click", selectNode)
        .call(d3.drag()
            .on("start", dragstarted)
            .on("drag", dragged)
            .on("end", dragended));

    textElements = g.append("g")
        .attr("class", "texts")
        .selectAll("text")
        .data(graph.nodes)
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
        //.attr("dx", 30)
        .attr("dy", 40) // vertical position of text
        .attr("text-anchor", "middle")

}

function ticked() {
    linkElements
        .attr("x1", function (d) {
            return d.source.x;
        })
        .attr("y1", function (d) {
            return d.source.y;
        })
        .attr("x2", function (d) {
            return d.target.x;
        })
        .attr("y2", function (d) {
            return d.target.y;
        });
    nodeElements
        .attr("cx", function (d) {
            //return d.x;
            return d.x = Math.max(radius, Math.min(width - radius, d.x));
        })
        .attr("cy", function (d) {
            //return d.y;
            return d.y = Math.max(radius, Math.min(height - radius, d.y));
        })
        .each(d => {
            d3.select('#t_' + d.id).attr('x', d.x + 10).attr('y', d.y + 3);
        });
    textElements
        .attr('x', function (d) {
            return d.x
        })
        .attr('y', function (d) {
            return d.y
        });
}

function selectNode(d) {
    console.log(d);
}


function dragstarted(d) {
    if (!d3.event.active) simulation.alphaTarget(0.3).restart();
    d.fx = d.x;
    d.fy = d.y;
}

function dragged(d) {
    d.fx = d3.event.x;
    d.fy = d3.event.y;
}

function dragended(d) {
    if (!d3.event.active) simulation.alphaTarget(0);
    d.fx = null;
    d.fy = null;
}